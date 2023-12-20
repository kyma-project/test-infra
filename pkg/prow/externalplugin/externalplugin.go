package externalplugin

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/kyma-project/test-infra/pkg/github/client/v2"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/test-infra/prow/config/secret"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp/externalplugins"
	"k8s.io/test-infra/prow/plugins"
)

const EventTypeField = "event-type"

// TODO: do we really need this interface? Maybe we can add AddFlags to CliOptions interface.
type ConfigOptionsGroup interface {
	AddFlags(fs *flag.FlagSet)
}

// Opts holds configuration for external plugin instance.
type Opts struct {
	Port              int
	Github            client.GithubClientConfig
	WebhookSecretPath string
	LogLevel          string
	DryRun            bool
}

// CliOptions is an interface to externalplugin cli flags.
type CliOptions interface {
	NewFlags() *flag.FlagSet
	ParseFlags(fs *flag.FlagSet)
	GetPort() int
	// TODO: Implement support for setting log level
	// GetLogLevel() string
}

// Event represent event passed to plugin handler.
// It's constructed from GitHub webhook.
type Event struct {
	EventType string
	EventGUID string
	Payload   []byte
}

// Plugin is an externaplugin instance object.
type Plugin struct {
	Name               string
	PluginsConfigAgent *plugins.ConfigAgent
	tokenGenerator     func() []byte
	handler            func(string, string, []byte)
	webhookHandlers    map[string]func(*Plugin, Event)
	// TODO: change logger type to logging.LoggerInterface
	logger *zap.SugaredLogger
}

// NewFlags create new flagset with default options. These options are common for all external plugins.
func (o *Opts) NewFlags() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.IntVar(&o.Port, "port", 8080, "Plugin port to listen on.")
	fs.BoolVar(&o.DryRun, "dry-run", false, "Run in dry-run mode - no actual changes will be made.")
	fs.StringVar(&o.WebhookSecretPath, "hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing GitHub HMAC secret")
	fs.StringVar(&o.LogLevel, "log-level", "info", "Set log level.")
	o.Github.AddFlags(fs)
	return fs
}

// ParseFlags parses cli arguments in to provided flagset.
func (o *Opts) ParseFlags(fs *flag.FlagSet) {
	fs.Parse(os.Args[1:])
}

// GetPort return port number for Plugin to listen on.
func (o *Opts) GetPort() int {
	return o.Port
}

// GetName return Plugin name.
func (p *Plugin) GetName() string {
	return p.Name
}

// NewGithubClient build GitHub client from provided GitHub options.
func NewGithubClient(githubOptions prowflagutil.GitHubOptions, dryRun bool) (github.Client, error) {
	return githubOptions.GitHubClient(dryRun)
}

// NewLogger return zap sugaredlogger with two output targets. All logs with severity Error or higher will be sent to stderr.
// All logs with severity lower than Error will be sed to stdout. This allows gcp logging correctly recognize log message severity.
// TODO: Use our console logger.
func NewLogger() *zap.SugaredLogger {
	errorMessage := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	infoMessage := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	consoleInfo := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, errorMessage),
		zapcore.NewCore(consoleEncoder, consoleInfo, infoMessage),
	)

	return zap.New(core).Sugar()
}

// WithLogger adds provided logger to the Plugin object.
func (p *Plugin) WithLogger(l *zap.SugaredLogger) *Plugin {
	p.logger = l
	return p
}

// WithWebhookSecret adds webhook secret path to the Prow secret agent.
func (p *Plugin) WithWebhookSecret(webhookSecretPath string) *Plugin {
	if err := secret.Add(webhookSecretPath); err != nil {
		p.logger.Errorw("Could not add path to secret agent.", "error", err.Error())
		return p
	}
	p.WithTokenGenerator(secret.GetTokenGenerator(webhookSecretPath))
	return p
}

// WithTokenGenerator sets custom tokenGenerator function if the default implementation can't be used.
func (p *Plugin) WithTokenGenerator(tg func() []byte) *Plugin {
	if tg != nil {
		p.tokenGenerator = tg
	}
	return p
}

// WithHandler sets custom handler function for GitHub event payload if the default implementation can't be used.
func (p *Plugin) WithHandler(handler func(string, string, []byte)) *Plugin {
	if handler != nil {
		p.handler = handler
	}
	return p
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, eventGUID, payload, ok, httpStatus := github.ValidateWebhook(w, r, p.tokenGenerator)
	if !ok {
		return
	}
	w.WriteHeader(httpStatus)
	p.handler(eventType, eventGUID, payload)
}

// RegisterWebhookHandler registers function that will be executed when Plugin receives it from GitHub.
// Only one function can be handled for each webhook.
func (p *Plugin) RegisterWebhookHandler(webhookName string, handler func(*Plugin, Event)) {
	// lazy init webhook map
	if p.webhookHandlers == nil {
		p.webhookHandlers = make(map[string]func(*Plugin, Event))
	}
	// add only if handler does not exist
	if _, ok := p.webhookHandlers[webhookName]; !ok {
		p.webhookHandlers[webhookName] = handler
	} else {
		p.logger.Warnw("Webhook handler already defined. Adding skipped.", "webhook", webhookName)
	}
}

// defaultHandler defines default handling function that is used when no other implementation is provided.
func (p *Plugin) defaultHandler(eventType, eventGUID string, payload []byte) {
	eventPayload := Event{
		EventType: eventType,
		EventGUID: eventGUID,
		Payload:   payload,
	}

	if wh, ok := p.webhookHandlers[eventType]; ok {
		go wh(p, eventPayload)
	} else {
		p.logger.Infow("skipping unsupported event", EventTypeField, eventType, github.EventGUID, eventGUID)
	}
}

// Start runs http server with provided plugin as a handler function provider.
// Start will set defaultHandler as a requests handler function if it's nil in provided plugin.
func Start(p *Plugin, helpProvider externalplugins.ExternalPluginHelpProvider, o CliOptions) {
	p.logger.With("plugin", p.GetName())
	if p.handler == nil {
		p.handler = p.defaultHandler
	}
	if p.tokenGenerator == nil {
		p.logger.Fatal("TokenGenerator cannot be empty.")
	}

	mux := http.NewServeMux()
	mux.Handle("/", p)
	externalplugins.ServeExternalPluginHelp(mux, logrus.StandardLogger().WithField("plugin", p.GetName()), helpProvider)

	s := http.Server{
		Addr:    ":" + strconv.Itoa(o.GetPort()),
		Handler: mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			p.logger.Fatalw("Plugin listen error.", "error", err.Error())
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	<-sig
	if err := s.Shutdown(ctx); err != nil {
		p.logger.Fatalw("Error closing server.", "error", err.Error())
	}
}
