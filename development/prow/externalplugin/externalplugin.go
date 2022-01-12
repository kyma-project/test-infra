package externalplugin

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/secret"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp/externalplugins"
	"k8s.io/test-infra/prow/plugins"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const EventTypeField = "event-type"

type ConfigOptionsGroup interface {
	AddFlags(fs *flag.FlagSet)
}

type Opts struct {
	Port              int
	Github            prowflagutil.GitHubOptions
	WebhookSecretPath string
	LogLevel          string
	DryRun            bool
}

type CliOptions interface {
	GatherDefaultOptions() *flag.FlagSet
	Parse(fs *flag.FlagSet)
	GetPort() int
	GetLogLevel() string
}

type Event struct {
	EventType string
	EventGUID string
	Payload   []byte
}

type Plugin struct {
	Name               string
	GitHub             github.Client
	PluginsConfigAgent *plugins.ConfigAgent

	tokenGenerator  func() []byte
	handler         func(string, string, []byte)
	webhookHandlers map[string]func(*Plugin, Event)
}

func (o *Opts) GatherDefaultOptions() *flag.FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.IntVar(&o.Port, "port", 8080, "Plugin port to listen on.")
	fs.BoolVar(&o.DryRun, "dry-run", false, "Run in dry-run mode - no actual changes will be made.")
	fs.StringVar(&o.WebhookSecretPath, "hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing GitHub HMAC secret")
	fs.StringVar(&o.LogLevel, "log-level", "info", "Set log level.")
	o.Github.AddFlags(fs)
	return fs
}

func (o *Opts) Parse(fs *flag.FlagSet) {
	fs.Parse(os.Args[1:])
}

func (o *Opts) GetPort() int {
	return o.Port
}

func (o *Opts) GetLogLevel() string {
	return o.LogLevel
}

func (p *Plugin) GetName() string {
	return p.Name
}

func (p *Plugin) WithGithubClient(githubOptions prowflagutil.GitHubOptions, dryRun bool) *Plugin {
	ghClient, err := githubOptions.GitHubClient(dryRun)
	if err != nil {
		logrus.WithError(err).Fatal("Could not get github client.")
	}
	p.GitHub = ghClient
	return p
}

// WithWebhookSecret initializes adds webhook secret path to the Prow secret agent.
func (p *Plugin) WithWebhookSecret(webhookSecretPath string) *Plugin {
	if err := secret.Add(webhookSecretPath); err != nil {
		logrus.WithError(err).Error("Could not add path to secret agent.")
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
	eventType, eventGUID, payload, ok, _ := github.ValidateWebhook(w, r, p.tokenGenerator)
	if !ok {
		return
	}
	p.handler(eventType, eventGUID, payload)
}

// HandleWebhook registers function that will be executed when Plugin receives it from GitHub.
// Only one function can be handled for each webhook.
func (p *Plugin) HandleWebhook(webhookName string, handler func(*Plugin, Event)) {
	// lazy init webhook map
	if p.webhookHandlers == nil {
		p.webhookHandlers = make(map[string]func(*Plugin, Event))
	}
	// add only if handler does not exist
	if _, ok := p.webhookHandlers[webhookName]; !ok {
		p.webhookHandlers[webhookName] = handler
	} else {
		logrus.WithField("webhook", webhookName).Warn("Webhook handler already defined. Adding skipped.")
	}
}

// defaultHandler defines default handling function that is used when no other implementation is provided.
func (p *Plugin) defaultHandler(eventType, eventGUID string, payload []byte) {
	eventPayload := Event{
		EventType: eventType,
		EventGUID: eventGUID,
		Payload:   payload,
	}

	l := logrus.WithFields(
		logrus.Fields{
			EventTypeField:   eventType,
			github.EventGUID: eventGUID,
		},
	)

	if wh, ok := p.webhookHandlers[eventType]; ok {
		go wh(p, eventPayload)
	} else {
		l.Debug("skipping unknown event")
	}
}

func Start(p *Plugin, helpProvider externalplugins.ExternalPluginHelpProvider, o CliOptions) {
	if p.handler == nil {
		p.handler = p.defaultHandler
	}
	if p.tokenGenerator == nil {
		logrus.Fatal("TokenGenerator cannot be empty.")
	}

	lvl, err := logrus.ParseLevel(o.GetLogLevel())
	if err != nil {
		logrus.WithError(err).Fatal("Could not parse log level.")
	}
	logrus.SetLevel(lvl)
	log := logrus.StandardLogger().WithField("plugin", p.GetName())

	mux := http.NewServeMux()
	mux.Handle("/", p)
	externalplugins.ServeExternalPluginHelp(mux, log, helpProvider)

	s := http.Server{
		Addr:    ":" + strconv.Itoa(o.GetPort()),
		Handler: mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			logrus.WithError(err).Fatal("Plugin listen error.")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case <-sig:
		if err := s.Shutdown(ctx); err != nil {
			logrus.WithError(err).Fatal("Error closing server.")
		}
	}
}
