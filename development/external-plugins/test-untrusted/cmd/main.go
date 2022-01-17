package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-project/test-infra/development/prow/externalplugin"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
)

const (
	PluginName = "test-untrusted"
)

type githubClient interface {
	AddLabelWithContext(ctx context.Context, org string, repo string, number int, label string) error
	CreateCommentWithContext(ctx context.Context, org, repo string, number int, comment string) error
}

func EventHandler(server *externalplugin.Plugin, event externalplugin.Event) {
	l := externalplugin.NewLogger()
	defer l.Sync()
	l = l.With(externalplugin.EventTypeField, event.EventType,
		github.EventGUID, event.EventGUID,
	)
	var pr github.PullRequestEvent
	if err := json.Unmarshal(event.Payload, &pr); err != nil {
		l.Errorw("Failed unmarshal json payload.", "error", err.Error())
	}
	l = l.With("pr-number", pr.Number,
		"pr-sender", pr.Sender.Login,
	)
	switch pr.Action {
	case github.PullRequestActionOpened:
		pr.GUID = event.EventGUID
		if pr.Sender.Login == "dependabot[bot]" {
			l.Info("Received pull request event for supported user.")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err := server.GitHub.(githubClient).AddLabelWithContext(ctx, pr.Repo.Owner.Login, pr.Repo.Name, pr.Number, "ok-to-test")
			if err != nil {
				l.Errorw("Failed label PR.", "error", err.Error())
			} else {
				l.Info("Labeled pr as trusted.")
			}
			err = server.GitHub.(githubClient).CreateCommentWithContext(ctx, pr.Repo.Owner.Login, pr.Repo.Name, pr.Number, "/test all")
			if err != nil {
				l.Errorw("Failed comment on PR.", "error", err.Error())
			} else {
				l.Info("Triggered all tests.")
			}
		} else {
			l.Info("Ignoring event triggered by not supported user.")
		}
	default:
		l.Infow("Ignoring unsupported pull request action.", "pr_action", pr.Action)
	}
}

func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "test-untrusted add ok-to-test label on a pull requests created by users from outside of an github organisation. It checks pr author against list of supported users.",
	}
	return ph, nil
}

func main() {
	var ghClient githubClient

	logger := externalplugin.NewLogger()
	defer logger.Sync()

	server := externalplugin.Plugin{}
	server.WithLogger(logger)

	cliOptions := externalplugin.Opts{}
	fs := cliOptions.GatherDefaultOptions()
	cliOptions.Parse(fs)

	ghClient = externalplugin.NewGithubClient(cliOptions.Github, cliOptions.DryRun, logger)

	server.Name = PluginName
	server.WithWebhookSecret(cliOptions.WebhookSecretPath)
	server.WithGithubClient(ghClient)
	server.HandleWebhook("pull_request", EventHandler)
	externalplugin.Start(&server, HelpProvider, &cliOptions)
}
