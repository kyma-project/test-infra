package main

import (
	"github.com/kyma-project/test-infra/development/github/pkg/git"
	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/pluginhelp"
)

const (
	PluginName = "automated-approver"
)

// Plugin help description. This is published on Prow status plugin catalog.
func helpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "The automerge-approver plugin approves PRs matching defined conditions.",
	}
	return ph, nil
}

func main() {
	var err error

	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()

	// Initialize configuration options for clients.
	pluginOptions := externalplugin.Opts{}
	gitOptions := git.ClientConfig{}

	// Add client and plugin cli flags.
	fs := pluginOptions.NewFlags()
	gitOptions.AddFlags(fs)
	pluginOptions.ParseFlags(fs)

	atom.SetLevel(pluginOptions.LogLevel)

	// Create github.com client.
	ghClient, err := pluginOptions.Github.NewGithubClient()
	if err != nil {
		logger.Fatalw("Failed creating GitHub client", "error", err)
		panic(err)
	}
	logger.Debug("github client ready")

	// Create git factory for github.com.
	gitClientFactory, err := gitOptions.NewClient(git.WithTokenPath(pluginOptions.Github.TokenPath), git.WithGithubClient(ghClient))
	if err != nil {
		logger.Fatalw("Failed creating git client", "error", err)
		panic(err)
	}
	logger.Debug("git client ready")

	hb := handlerBackend{
		ghc:      ghClient,
		gcf:      gitClientFactory,
		logLevel: pluginOptions.LogLevel,
	}

	// Create and start plugin instance.
	server := externalplugin.Plugin{}
	server.WithLogger(logger)
	server.Name = PluginName
	server.WithWebhookSecret(pluginOptions.WebhookSecretPath)
	server.RegisterWebhookHandler("pull_request", hb.pullRequestEventHandler)
	externalplugin.Start(&server, helpProvider, &pluginOptions)
}
