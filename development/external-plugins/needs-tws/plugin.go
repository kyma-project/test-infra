package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/repoowners"
	"net/http"
	"regexp"
	"strings"
)

const (
	DefaultNeedsTwsLabel = "needs-tws-review"
	Reviewer             = "adamwalach"
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
	RequestReview(org, repo string, number int, logins []string) error
	CreateComment(org, repo string, number int, comment string) error
	IsCollaborator(org, repo, user string) (bool, error)
	AssignIssue(org, repo string, number int, logins []string) error
}

type reviewCtx struct {
	author, issueAuthor, body, htmlURL string
	number                             int
	repo                               github.Repo
	assignees                          []github.User
	approved                           bool
}

type Plugin struct {
	botUser        *github.UserData
	tokenGenerator func() []byte
	ghc            githubClient
	owc            repoowners.Interface
}

func (p Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, eventGUID, payload, ok, _ := github.ValidateWebhook(w, r, p.tokenGenerator)
	if !ok {
		return
	}
	if err := p.handleEvent(eventType, eventGUID, payload); err != nil {
		//error parsing event
	}
}

func (p *Plugin) handleEvent(eventType, eventGUID string, payload []byte) error {
	l := logrus.WithFields(logrus.Fields{
		"event-type":     eventType,
		github.EventGUID: eventGUID,
	})
	switch eventType {
	case "pull_request":
		l.Debug("Got pull_request event.")
		var pre github.PullRequestEvent
		if err := json.Unmarshal(payload, &pre); err != nil {
			return err
		}
		go func() {
			if err := p.handlePullRequest(l, pre); err != nil {
				l.WithError(err).Error("Failed to process Pull Request.")
			}
		}()
	case "pull_request_review":
		l.Debug("Got pull_request_review event")
		var re github.ReviewEvent
		if err := json.Unmarshal(payload, &re); err != nil {
			return err
		}
		go func() {
			if err := p.handlePullRequestReview(l, re); err != nil {
				l.WithError(err).Error("Failed to process Pull Request review.")
			}
		}()
	default:
		logrus.Debugf("skipping event of type: %q", eventType)
	}
	return nil
}

func (p *Plugin) handlePullRequest(l *logrus.Entry, e github.PullRequestEvent) error {
	if e.Action == github.PullRequestActionClosed || e.Action == github.PullRequestActionLabeled {
		return nil
	}
	pr := e.PullRequest
	if pr.Draft || pr.Merged {
		return nil
	}

	org := e.Repo.Owner.Login
	repo := e.Repo.Name
	number := e.Number

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
		l.Infof("Adding label to %s/%s#%d.", org, repo, number)
		return p.ghc.AddLabel(org, repo, number, DefaultNeedsTwsLabel)
	}
	return nil
}

func (p Plugin) handlePullRequestReview(l *logrus.Entry, re github.ReviewEvent) error {
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

func (p Plugin) handle(l *logrus.Entry, rc reviewCtx) error {
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

	// TODO fetch list of reviewers from fixed file from repo or config
	isTWS := author == Reviewer
	if !isTWS {
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
		l.WithError(err).Error("Could not determine if author is a collaborator.")
		return err
	}

	// User must be a collaborator
	if !isCollaborator && !isAuthor {
		return nil
	}

	if !isAuthor && !isAssignee {
		l.Infof("Assign PR #%v to %s", number, author)
		if err := p.ghc.AssignIssue(org, repoName, number, []string{author}); err != nil {
			l.WithError(err).Error("Failed to assign %s/%s#%d to %s", org, repoName, number, author)
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

func (p Plugin) hasMarkdownChanges(org, repo, sha string) (bool, error) {
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
