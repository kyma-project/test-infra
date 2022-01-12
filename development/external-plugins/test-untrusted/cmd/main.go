package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-project/test-infra/development/prow/externalplugin"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
)

const (
	PluginName = "test-untrusted"
)

func EventHandler(s *externalplugin.Plugin, e externalplugin.Event) {
	var event externalplugin.Event
	l := logrus.WithFields(
		logrus.Fields{
			externalplugin.EventTypeField: event.EventType,
			github.EventGUID:              event.EventGUID,
		},
	)
	var pr github.PullRequestEvent
	if err := json.Unmarshal(event.Payload, &pr); err != nil {
		l.WithError(err).Error("Failed unmarshal json payload.")
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
			l.WithField("severity", "INFO").Info("Received pull request event for supported user.")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err := s.GitHub.CreateCommentWithContext(ctx, pr.Repo.Owner.Login, pr.Repo.Name, pr.Number, "/test all")
			if err != nil {
				l.WithError(err).Error("Failed comment on PR.")
			} else {
				l.WithField("severity", "INFO").Info("Send command to run all tests.")
			}
		} else {
			l.WithField("severity", "INFO").Info("Event triggered by not supported user, ignoring.")
		}
	default:
		l.WithField("pr_action", pr.Action).WithField("severity", "INFO").Info("Ignoring unsupported pull request action.")
	}
}

func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "test-untrusted trigger all tests on pull requests created by users from outside of an github organisation. It checks pr author against list of supported users.",
	}
	return ph, nil
}

func main() {
	cliOptions := externalplugin.Opts{}
	fs := cliOptions.GatherDefaultOptions()
	cliOptions.Parse(fs)

	server := externalplugin.Plugin{}
	server.Name = PluginName
	server.WithWebhookSecret(cliOptions.WebhookSecretPath)
	server.WithGithubClient(cliOptions.Github, cliOptions.DryRun)
	server.HandleWebhook("pull_request", EventHandler)
	externalplugin.Start(&server, HelpProvider, &cliOptions)
}
