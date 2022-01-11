package prow

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/secret"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp/externalplugins"
	"k8s.io/test-infra/prow/plugins"
)

const EventTypeField = "event-type"

type ConfigOptionsGroup interface {
	AddFlags(fs *flag.FlagSet)
}

type GithubClient interface{}

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

type PluginEvent interface{}

type Plugin interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	GetName() string
}

type Event struct {
	EventType      string
	EventGUID      string
	Payload        []byte
	OK             bool
	HttpStatusCode int
}

type Server struct {
	Name               string
	TokenGenerator     func() []byte
	ValidateWebhook    func(http.ResponseWriter, *http.Request) (string, string, []byte, bool, int)
	GithubClient       GithubClient
	PluginsConfigAgent *plugins.ConfigAgent
	EventMux           func(chan interface{}, *Server)
	Handler            func(http.ResponseWriter, *http.Request)
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

func (s *Server) GetName() string {
	return s.Name
}

func (s *Server) WithValidateWebhook() *Server {
	if s.TokenGenerator == nil {
		logrus.Panic()
	}
	s.ValidateWebhook = func(w http.ResponseWriter, r *http.Request) (string, string, []byte, bool, int) {
		return github.ValidateWebhook(w, r, s.TokenGenerator)
	}
	return s
}

func (s *Server) WithGithubClient(githubOptions prowflagutil.GitHubOptions, dryRun bool) *Server {
	ghClient, err := githubOptions.GitHubClient(dryRun)
	if err != nil {
		logrus.WithError(err).Fatal("Could not get github client.")
	}
	s.GithubClient = ghClient
	return s
}

func (s *Server) WithTokenGenerator(webhookSecretPath string) *Server {
	if err := secret.Add(webhookSecretPath); err != nil {
		logrus.WithError(err).Fatal("Could not start secret agent.")
	}
	s.TokenGenerator = secret.GetTokenGenerator(webhookSecretPath)
	return s
}

func (s *Server) WithEventMux(eventMux func(chan interface{}, *Server)) *Server {
	s.EventMux = eventMux
	return s
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (s *Server) WithHandler() *Server {
	if s.ValidateWebhook == nil {

	}
	s.Handler = func(w http.ResponseWriter, r *http.Request) {
		eventType, eventGUID, payload, ok, httpStatusCode := s.ValidateWebhook(w, r)
		eventPayload := Event{
			EventType:      eventType,
			EventGUID:      eventGUID,
			Payload:        payload,
			OK:             ok,
			HttpStatusCode: httpStatusCode,
		}

		if !ok {
			return
		}
		l := logrus.WithFields(
			logrus.Fields{
				EventTypeField:   eventType,
				github.EventGUID: eventGUID,
			},
		)
		l.Info("Event received. Have a nice day.")
		c := make(chan interface{})
		go s.EventMux(c, s)
		c <- eventPayload

		eventMuxResponse := <-c
		if eventMuxResponse != nil {
			//	l.WithError(err).Error("Error parsing event.")
		}

	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler(w, r)
}

func Start(plugin Plugin, helpProvider externalplugins.ExternalPluginHelpProvider, o CliOptions) {
	lvl, err := logrus.ParseLevel(o.GetLogLevel())
	if err != nil {
		logrus.WithError(err).Fatal("Could not parse log level.")
	}
	logrus.SetLevel(lvl)
	log := logrus.StandardLogger().WithField("plugin", plugin.GetName())

	mux := http.NewServeMux()
	mux.Handle("/", plugin)
	externalplugins.ServeExternalPluginHelp(mux, log, helpProvider)

	s := http.Server{
		Addr:    ":" + strconv.Itoa(o.GetPort()),
		Handler: mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			logrus.WithError(err).Fatal("Server listen error.")
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
