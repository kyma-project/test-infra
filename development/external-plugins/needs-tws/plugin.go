package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"go.uber.org/zap"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/repoowners"
)

const (
	DefaultNeedsTwsLabel         = "do-not-merge/missing-docs-review"
	DefaultTechnicalWritersGroup = "technical-writers"
)

var markdownRe = regexp.MustCompile(".*.md")

func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "needs-tws checks if the Pull Request has modified Markdown files and blocks the merge until it is reviewed and approved by one of the Technical Writers.",
	}
	return ph, nil
}

type githubClient interface {
	GetSingleCommit(org, repo, SHA string) (github.RepositoryCommit, error)
	AddLabel(org, repo string, number int, label string) error
	RemoveLabel(org, repo string, number int, label string) error
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	CreateComment(org, repo string, number int, comment string) error
	IsCollaborator(org, repo, user string) (bool, error)
	AssignIssue(org, repo string, number int, logins []string) error
}

type ownersAliases interface {
	LoadOwnersAliases(l *zap.SugaredLogger, basedir, filename string) (repoowners.RepoAliases, error)
}

type AliasesClient struct {
	ownersAliases
}

func (o AliasesClient) LoadOwnersAliases(l *zap.SugaredLogger, basedir, filename string) (repoowners.RepoAliases, error) {
	l.Debug("Load OWNERS_ALIASES")
	path := filepath.Join(basedir, filename)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(path)
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
	botUser        *github.UserData
	tokenGenerator func() []byte
	ghc            githubClient
	gcf            git.ClientFactory
	oac            ownersAliases
}

func (p *PluginBackend) HandlePR(ep *externalplugin.Plugin, e externalplugin.Event) {
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
		l.Errorw("Failed unmarshal json payload.", "error", err.Error())
		return //return prematurely
	}
	if err := p.handlePullRequest(l, pre); err != nil {
		l.Errorw("Failed to handle PR event", "error", err.Error())
	}
}

func (p *PluginBackend) HandlePRReview(ep *externalplugin.Plugin, e externalplugin.Event) {
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
		l.Errorw("Failed unmarshal json payload.", "error", err.Error())
		return //return prematurely
	}
	if err := p.handlePullRequestReview(l, re); err != nil {
		l.Errorw("Failed to handle Pull Request Review event", "error", err.Error())
	}
}

func (p *PluginBackend) handlePullRequest(l *zap.SugaredLogger, e github.PullRequestEvent) error {
	if e.Action == github.PullRequestActionClosed || e.Action == github.PullRequestActionLabeled {
		return nil
	}
	pr := e.PullRequest
	if pr.Draft || pr.Merged {
		return nil
	}

	org := e.Repo.Owner.Login
	repo := e.Repo.Name
	number := e.PullRequest.Number

	labels, err := p.ghc.GetIssueLabels(org, repo, number)
	if err != nil {
		return err
	}
	if github.HasLabel(DefaultNeedsTwsLabel, labels) {
		l.Debug("Pull request already has a label.")
		return nil
	}

	hasChanges, err := p.hasMarkdownChanges(org, repo, pr.Head.SHA)
	if err != nil {
		return err
	}
	if hasChanges {
		l.Debugf("Adding label to %s/%s#%d.", org, repo, number)
		return p.ghc.AddLabel(org, repo, number, DefaultNeedsTwsLabel)
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
	return p.handle(l, rc)
}

func (p PluginBackend) handle(l *zap.SugaredLogger, rc reviewCtx) error {
	author := rc.author
	issueAuthor := rc.issueAuthor
	org := rc.repo.Owner.Login
	repoName := rc.repo.Name
	assignees := rc.assignees
	number := rc.number
	//body := rc.body
	approved := rc.approved

	isAuthor := author == issueAuthor
	// author cannot review their own PR
	if isAuthor {
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
	twsGroup := repoAliases[DefaultTechnicalWritersGroup]
	if !twsGroup.Has(author) {
		l.Infof("'%s' is not in the group of required reviewers.", author)
		return nil
	}

	var isAssignee bool
	for _, a := range assignees {
		if a.Login == author {
			isAssignee = true
			break
		}
	}

	isCollaborator, err := p.ghc.IsCollaborator(org, repoName, author)
	if err != nil {
		l.Errorw("Could not determine if author is a collaborator.", "error", err)
		return err
	}

	// User must be a collaborator
	if !isCollaborator && !isAuthor {
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
		if err := p.ghc.RemoveLabel(org, repoName, number, DefaultNeedsTwsLabel); err != nil {
			return err
		}
	}

	return nil
}

func (p PluginBackend) hasMarkdownChanges(org, repo, sha string) (bool, error) {
	commit, err := p.ghc.GetSingleCommit(org, repo, sha)
	if err != nil {
		return false, err
	}
	for _, f := range commit.Files {
		if markdownRe.MatchString(f.Filename) {
			return true, nil
		}
	}
	return false, nil
}
