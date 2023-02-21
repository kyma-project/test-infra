package main

import (
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

	// Add client and plugin cli flags.
	fs := pluginOptions.NewFlags()
	pluginOptions.ParseFlags(fs)

	atom.SetLevel(pluginOptions.LogLevel)

	// Create github.com client.
	ghClient, err := pluginOptions.Github.NewGithubClient()
	if err != nil {
		logger.Fatalw("Failed creating GitHub client", "error", err)
		panic(err)
	}
	logger.Debug("github client ready")

	hb := handlerBackend{
		ghc:      ghClient,
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
