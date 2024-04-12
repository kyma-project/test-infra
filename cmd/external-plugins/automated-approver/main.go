package main

import (
	consolelog "github.com/kyma-project/test-infra/pkg/logging"
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"sigs.k8s.io/prow/prow/config"
	"sigs.k8s.io/prow/prow/pluginhelp"
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

	hb := HandlerBackend{}

	// Initialize PR locks.
	hb.PrLocks = make(map[string]map[string]map[int]map[string]context.CancelFunc)

	// Add client and plugin cli flags.
	fs := pluginOptions.NewFlags()
	fs.StringVar(&hb.RulesPath, "rules-path", "", "Path to the configuration file.")
	fs.IntVar(&hb.WaitForStatusesTimeout, "wait-for-statuses-timeout", 30, "Timeout in seconds for waiting for statuses.")
	fs.IntVar(&hb.WaitForContextsCreationTimeout, "wait-for-contexts-creation-timeout", 30, "Timeout in seconds for waiting for contexts creation.")
	pluginOptions.ParseFlags(fs)

	atom.UnmarshalText([]byte(pluginOptions.LogLevel))
	level, err := zapcore.ParseLevel(pluginOptions.LogLevel)
	if err != nil {
		logger.Fatalw("Failed parsing log level", "error", err)
		panic(err)
	}
	hb.LogLevel = level

	// Create GitHub.com client.
	ghClient, err := pluginOptions.Github.NewGithubClient()
	if err != nil {
		logger.Fatalw("Failed creating GitHub client", "error", err)
		panic(err)
	}
	logger.Info("github client ready")
	hb.Ghc = ghClient

	err = hb.ReadConfig()
	if err != nil {
		logger.Fatalw("Failed reading config", "error", err)
		panic(err)
	}
	logger.Debugf("config: %+v", hb.Conditions)
	logger.Info("config ready")

	// Watch hb.rulesPath for changes and reload configured rules.
	go hb.WatchConfig(logger)

	// Create and start plugin instance.
	server := externalplugin.Plugin{}
	server.WithLogger(logger)
	server.Name = PluginName
	server.WithWebhookSecret(pluginOptions.WebhookSecretPath)
	server.RegisterWebhookHandler("pull_request", hb.PullRequestEventHandler)
	server.RegisterWebhookHandler("pull_request_review", hb.PullRequestReviewEventHandler)
	externalplugin.Start(&server, helpProvider, &pluginOptions)
}
