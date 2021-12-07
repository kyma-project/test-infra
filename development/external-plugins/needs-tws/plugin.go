package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"net/http"
	"regexp"
)

const (
	DefaultNeedsTwsLabel = "needs-tws-review"
)

var MarkdownRegexp = regexp.MustCompile(".*.md")

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
	RequestReview(org, repo string, number int, logins []string) error
	CreateComment(org, repo string, number int, comment string) error
}

type Plugin struct {
	botUser        *github.UserData
	tokenGenerator func() []byte
	ghc            githubClient
}

func (s Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, eventGUID, payload, ok, _ := github.ValidateWebhook(w, r, s.tokenGenerator)
	if !ok {
		return
	}
	if err := s.handleEvent(eventType, eventGUID, payload); err != nil {
		//error parsing event
	}
}

func (s *Plugin) handleEvent(eventType, eventGUID string, payload []byte) error {
	l := logrus.WithFields(logrus.Fields{
		"event-type":     eventType,
		github.EventGUID: eventGUID,
	})
	switch eventType {
	case "pull_request":
		var pr github.PullRequestEvent
		if err := json.Unmarshal(payload, &pr); err != nil {
			return err
		}
		go func() {
			if err := s.handlePullRequest(l, pr); err != nil {
				logrus.WithError(err).WithFields(l.Data).Info("Failed to check PR for TWs approval.")
			}
		}()
	default:
		logrus.Debugf("skipping event of type: %q", eventType)
	}
	return nil
}

func (s *Plugin) handlePullRequest(l *logrus.Entry, e github.PullRequestEvent) error {
	if e.Action != github.PullRequestActionClosed && e.Action != github.PullRequestActionLabeled {
		return nil
	}
	pr := e.PullRequest
	if pr.Draft || pr.Merged || pr.MergeSHA != nil {
		return nil
	}

	org := e.Repo.Owner.Login
	repo := e.Repo.Name
	number := e.Number

	changes, err := s.ghc.GetPullRequestChanges(org, repo, number)
	if err != nil {
		return err
	}
	labels, err := s.ghc.GetIssueLabels(org, repo, number)
	if err != nil {
		return err
	}

	var labelled bool
	for _, label := range labels {
		if label.Name == DefaultNeedsTwsLabel {
			labelled = true
			break
		}
	}
	mdChanged := hasMarkdownChanges(changes)

	if mdChanged && !labelled {
		return s.ghc.AddLabel(org, repo, number, DefaultNeedsTwsLabel)
	}
	if !mdChanged && labelled {
		return s.ghc.RemoveLabel(org, repo, number, DefaultNeedsTwsLabel)
	}
	return nil
}

func hasMarkdownChanges(changes []github.PullRequestChange) bool {
	for _, c := range changes {
		if MarkdownRegexp.MatchString(c.Filename) {
			return true
		}
	}
	return false
}
