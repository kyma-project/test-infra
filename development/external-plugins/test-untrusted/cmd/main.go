package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-project/test-infra/development/prow"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
)

const (
	PluginName = "test-untrusted"
)

type GithubClient interface {
	CreateComment(org, repo string, number int, comment string) error
	CreateCommentWithContext(ctx context.Context, org, repo string, number int, comment string) error
}

func EventMux(c chan interface{}, s *prow.Server) {
	var event prow.Event
	e := <-c
	event = e.(prow.Event)
	l := logrus.WithFields(
		logrus.Fields{
			prow.EventTypeField: event.EventType,
			github.EventGUID:    event.EventGUID,
		},
	)
	switch event.EventType {
	case "pull_request":
		var pr github.PullRequestEvent
		if err := json.Unmarshal(event.Payload, &pr); err != nil {
			c <- err
		}
		l = l.WithFields(
			logrus.Fields{
				"pr-number": pr.Number,
				"pr-sender": pr.Sender.Login,
			})
		switch pr.Action {
		case github.PullRequestActionOpened, github.PullRequestActionReopened, github.PullRequestActionSynchronize:
			pr.GUID = event.EventGUID
			if pr.Sender.Login == "dependabot[bot]" {
				l.Info("Received pull request event for supported user.")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				err := s.GithubClient.(GithubClient).CreateCommentWithContext(ctx, pr.Repo.Owner.Login, pr.Repo.Name, pr.Number, "/test all")
				if err != nil {
					l.WithError(err).Error("Failed comment on PR.")
				} else {
					l.Info("Send command to run all tests.")
				}
			} else {
				l.Info("Event triggered by not supported user, ignoring.")
			}
		default:
			l.WithField("pr_action", pr.Action).Info("Ignoring unsupported pull request action.")
		}
	default:
		l.Info("Ignoring unsupported event type.")
	}
	c <- nil
}

func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "needs-tws checks if the Pull Request has modified Markdown files and blocks the merge until it is reviewed and approved by one of the Technical Writers.",
	}
	return ph, nil
}

func main() {
	cliOptions := prow.Opts{}
	fs := cliOptions.GatherDefaultOptions()
	cliOptions.Parse(fs)

	server := prow.Server{}
	server.Name = PluginName
	server.WithTokenGenerator(cliOptions.WebhookSecretPath).WithValidateWebhook()
	server.WithGithubClient(cliOptions.Github, cliOptions.DryRun)
	server.WithEventMux(EventMux).WithHandler()
	prow.Start(&server, HelpProvider, &cliOptions)
}
