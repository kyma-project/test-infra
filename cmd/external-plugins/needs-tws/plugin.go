package main

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"sigs.k8s.io/prow/prow/config"
	"sigs.k8s.io/prow/prow/git/v2"
	"sigs.k8s.io/prow/prow/github"
	"sigs.k8s.io/prow/prow/pluginhelp"
	"sigs.k8s.io/prow/prow/repoowners"
)

const (
	DefaultNeedsTwsLabel         = "do-not-merge/missing-docs-review"
	DefaultTechnicalWritersGroup = "technical-writers"
)

var markdownRe = regexp.MustCompile(`^.*\.md$`)

func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "needs-tws checks if the Pull Request has modified Markdown files and blocks the merge until it is reviewed and approved by one of the Technical Writers.",
	}
	return ph, nil
}

type githubClient interface {
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
	AddLabel(org, repo string, number int, label string) error
	RemoveLabel(org, repo string, number int, label string) error
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	CreateComment(org, repo string, number int, comment string) error
	IsCollaborator(org, repo, user string) (bool, error)
	AssignIssue(org, repo string, number int, logins []string) error
	BotUserChecker() (func(string) bool, error)
}

type ownersAliases interface {
	LoadOwnersAliases(l *zap.SugaredLogger, basedir, filename string) (repoowners.RepoAliases, error)
}

type AliasesClient struct {
}

func (o AliasesClient) LoadOwnersAliases(l *zap.SugaredLogger, basedir, filename string) (repoowners.RepoAliases, error) {
	l.Debug("Load OWNERS_ALIASES")
	path := filepath.Join(basedir, filename)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return repoowners.ParseAliasesConfig(b)
}

type reviewCtx struct {
	author, issueAuthor, body, htmlURL string
	number                             int
	repo                               github.Repo
	assignees                          []github.User
	approved                           bool
}

type PluginBackend struct {
	ghc githubClient
	gcf git.ClientFactory
	oac ownersAliases
}

// PullRequestHandler handles all pull_request events from GitHub.
func (p *PluginBackend) PullRequestHandler(_ *externalplugin.Plugin, e externalplugin.Event) {
	eventGUID := e.EventGUID
	eventType := e.EventType
	payload := e.Payload

	l := externalplugin.NewLogger().With(
		externalplugin.EventTypeField, eventType,
		github.EventGUID, eventGUID,
	)
	defer l.Sync()

	l.Debug("Got pull_request event")
	var pre github.PullRequestEvent
	if err := json.Unmarshal(payload, &pre); err != nil {
		l.Errorw("Failed unmarshal json payload.", "error", err)
		return //return prematurely
	}
	if err := p.handlePullRequest(l, pre); err != nil {
		l.Errorw("Failed to handleReview PR event", "error", err)
	}
}

// PullRequestReviewHandler handles all pull_request_review events from GitHub
func (p *PluginBackend) PullRequestReviewHandler(_ *externalplugin.Plugin, e externalplugin.Event) {
	eventGUID := e.EventGUID
	eventType := e.EventType
	payload := e.Payload

	l := externalplugin.NewLogger().With(
		externalplugin.EventTypeField, eventType,
		github.EventGUID, eventGUID,
	)
	defer l.Sync()

	l.Debug("Got pull_request_review event")
	var re github.ReviewEvent
	if err := json.Unmarshal(payload, &re); err != nil {
		l.Errorw("Failed unmarshal json payload.", "error", err)
		return //return prematurely
	}
	if err := p.handlePullRequestReview(l, re); err != nil {
		l.Errorw("Failed to handleReview Pull Request Review event", "error", err)
	}
}

func (p *PluginBackend) handlePullRequest(l *zap.SugaredLogger, e github.PullRequestEvent) error {
	if e.Action == github.PullRequestActionClosed {
		// Pull Request is closed. There's no need to handle it at all.
		return nil
	}
	pr := e.PullRequest
	org := e.Repo.Owner.Login
	repo := e.Repo.Name
	number := pr.Number
	// TODO (@Ressetkk): Configure this based on per-repository config
	twsLabel := DefaultNeedsTwsLabel

	switch e.Action {
	case github.PullRequestActionOpened, github.PullRequestActionReopened, github.PullRequestActionSynchronize:
		var changed bool

		changes, err := p.ghc.GetPullRequestChanges(org, repo, number)
		if err != nil {
			return err
		}
		for _, c := range changes {
			if markdownRe.MatchString(strings.ToLower(c.Filename)) {
				changed = true
				break
			}
		}
		labels, err := p.ghc.GetIssueLabels(org, repo, number)
		if err != nil {
			return err
		}
		hasLabel := github.HasLabel(twsLabel, labels)

		if !changed {
			l.Debugf("Files not changed in %s/%s#%d@%s", org, repo, number, pr.Head.SHA)
			if hasLabel {
				l.Debug("remove stale label")
				return p.ghc.RemoveLabel(org, repo, number, twsLabel)
			}
			return nil
		}
		if hasLabel {
			//we do not need to add another label
			return nil
		}
		return p.ghc.AddLabel(org, repo, number, twsLabel)
	case github.PullRequestActionUnlabeled:
		bc, err := p.ghc.BotUserChecker()
		if err != nil {
			return err
		}
		isBot := bc(e.Sender.Login)
		if isBot {
			// label removed by a bot. Skip.
			return nil
		}
		deleted := e.Label.Name == twsLabel
		if !deleted {
			// not a missing-docs label. Skip.
			return nil
		}
		sender := e.Sender.Login
		isCollaborator, err := p.ghc.IsCollaborator(org, repo, sender)
		if err != nil {
			return nil
		}
		if !isCollaborator {
			// re-add label and notice only collaborators can bypass documentation review
			err := p.ghc.AddLabel(org, repo, number, twsLabel)
			if err != nil {
				l.Errorw("Failed to re-add label.", "error", err)
				return err
			}
			return p.ghc.CreateComment(org, repo, number, fmt.Sprintf("@%s, you cannot bypass documentation review. Only team members with write access are allowed to do it.", sender))
		}
		return p.ghc.CreateComment(org, repo, number, fmt.Sprintf("Documentation review manually bypassed by @%s.", sender))
	}
	return nil
}

func (p PluginBackend) handlePullRequestReview(l *zap.SugaredLogger, re github.ReviewEvent) error {
	if re.Action != github.ReviewActionSubmitted {
		return nil
	}
	rc := reviewCtx{
		author:      re.Review.User.Login,
		issueAuthor: re.PullRequest.User.Login,
		repo:        re.Repo,
		body:        re.Review.Body,
		htmlURL:     re.Repo.HTMLURL,
		number:      re.PullRequest.Number,
		assignees:   re.PullRequest.Assignees,
	}
	reviewState := github.ReviewState(strings.ToUpper(string(re.Review.State)))
	if reviewState == github.ReviewStateApproved {
		rc.approved = true
	} else if reviewState == github.ReviewStateChangesRequested {
		rc.approved = false
	} else {
		return nil
	}
	return p.handleReview(l, rc)
}

func (p PluginBackend) handleReview(l *zap.SugaredLogger, rc reviewCtx) error {
	// normalize GitHub names by making them lowercase and strip any unnecessary prefixes
	author := github.NormLogin(rc.author)
	issueAuthor := github.NormLogin(rc.issueAuthor)
	org := rc.repo.Owner.Login
	repoName := rc.repo.Name
	assignees := rc.assignees
	number := rc.number
	//body := rc.body
	approved := rc.approved

	isAuthor := author == issueAuthor

	if isAuthor {
		// Author cannot review their own PR. Let's ignore this event.
		return nil
	}

	gc, err := p.gcf.ClientFor(org, repoName)
	if err != nil {
		l.Errorw("Could not initialize git client.", "error", err)
		return err
	}
	repoAliases, err := p.oac.LoadOwnersAliases(l, gc.Directory(), "OWNERS_ALIASES")
	if err != nil {
		l.Errorw("Could not fetch owners aliases.", "error", err)
	}
	twsGroup := repoAliases.ExpandAlias(DefaultTechnicalWritersGroup)
	if !twsGroup.Has(author) {
		// User is not a member of TWs group defined in OWNERS_ALIASES.
		// possibly review meant for something else. Let's ignore this event.
		return nil
	}

	var isAssignee bool
	for _, a := range assignees {
		if github.NormLogin(a.Login) == author {
			isAssignee = true
			break
		}
	}

	isCollaborator, err := p.ghc.IsCollaborator(org, repoName, author)
	if err != nil {
		l.Errorw("Could not determine if author is a collaborator.", "error", err)
		return err
	}

	if !isCollaborator && !isAuthor {
		// User that left a review is not a collaborator.
		// Maybe they were removed from the team members, or they have never been in there.
		// Ignore this event.
		return nil
	}

	if !isAuthor && !isAssignee {
		l.Infof("Assign PR #%v to %s", number, author)
		if err := p.ghc.AssignIssue(org, repoName, number, []string{author}); err != nil {
			l.With("error", err).Errorf("Failed to assign %s/%s#%d to %s", org, repoName, number, author)
		}
	}

	// manipulate the label based once checked all the assumptions
	labels, err := p.ghc.GetIssueLabels(org, repoName, number)
	if err != nil {
		return err
	}
	hasLabel := github.HasLabel(DefaultNeedsTwsLabel, labels)
	if hasLabel && approved {
		// remove label when PR is approved and has labels.
		l.Infof("Remove label from %s/%s#%d.", org, repoName, number)
		return p.ghc.RemoveLabel(org, repoName, number, DefaultNeedsTwsLabel)
	}
	if !hasLabel && !approved {
		// re-add label if PR is not approved.
		l.Infof("Add label to %s/%s#%d", org, repoName, number)
		return p.ghc.AddLabel(org, repoName, number, DefaultNeedsTwsLabel)
	}
	return nil
}
