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

var ghClient githubClient

// EventHandler handles PullRequest opened events.
// It will label PR with ok-to-test to make PR trusted and comment with /test all to trigger tests for opened action.
func EventHandler(_ *externalplugin.Plugin, event externalplugin.Event) {
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
		if pr.Sender.Login == "dependabot[bot]" || pr.Sender.Login == "neighbors-dev-bot" {
			l.Info("Received pull request event for supported user.")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			err := ghClient.AddLabelWithContext(ctx, pr.Repo.Owner.Login, pr.Repo.Name, pr.Number, "ok-to-test")
			if err != nil {
				l.Errorw("Failed label PR.", "error", err.Error())
			} else {
				l.Info("Labeled pr as trusted.")
			}
			err = ghClient.CreateCommentWithContext(ctx, pr.Repo.Owner.Login, pr.Repo.Name, pr.Number, "/test all")
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

// HelpProvider provides plugin help description.
func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "test-untrusted add ok-to-test label on a opened pull request to make it trusted. It will comment with /test all to run tests for opened action. It acts for PRs created by users from outside of an github organisation but for which we need automatically test PRs. It acts for pr authors present on list of allowed users.",
	}
	return ph, nil
}

func main() {
	var err error
	logger := externalplugin.NewLogger()
	defer logger.Sync()

	server := externalplugin.Plugin{}
	server.WithLogger(logger)

	cliOptions := externalplugin.Opts{}
	fs := cliOptions.NewFlags()
	cliOptions.ParseFlags(fs)

	ghClient, err = externalplugin.NewGithubClient(cliOptions.Github.GitHubOptions, cliOptions.DryRun)
	if err != nil {
		logger.Fatal("Could not get github client.")
	}

	server.Name = PluginName
	server.WithWebhookSecret(cliOptions.WebhookSecretPath)
	server.RegisterWebhookHandler("pull_request", EventHandler)
	externalplugin.Start(&server, HelpProvider, &cliOptions)
}
