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
)

const EventTypeField = "event-type"

type ConfigOptionsGroup interface {
	AddFlags(fs *flag.FlagSet)
}

type CliOptions interface {
	GatherDefaultOptions() *flag.FlagSet
	Parse(fs *flag.FlagSet)
}

type GithubClient interface{}

type Plugin interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	GetName() string
}

type EventMux interface{}

type Opts struct {
	Port              int
	Github            prowflagutil.GitHubOptions
	WebhookSecretPath string
	LogLevel          string
	DryRun            bool
}

type Server struct {
	Name           string
	TokenGenerator func() []byte
	DemuxEvent     func(eventType, eventGUID string, payload []byte) error
	GithubClient   GithubClient
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

func (s *Server) GetName() string {
	return s.Name
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, eventGUID, payload, ok, _ := github.ValidateWebhook(w, r, s.TokenGenerator)
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

	if err := s.DemuxEvent(eventType, eventGUID, payload); err != nil {
		l.WithError(err).Error("Error parsing event.")
	}
}

func (s *Server) WithTokenGenerator(webhookSecretPath string) Plugin {
	s.TokenGenerator = secret.GetTokenGenerator(webhookSecretPath)
	return s
}

func (s *Server) WithGithubClient(githubOptions prowflagutil.GitHubOptions, dryRun bool) Plugin {
	ghClient, err := githubOptions.GitHubClient(dryRun)
	if err != nil {
		logrus.WithError(err).Fatal("Could not get github client.")
	}
	s.GithubClient = ghClient
	return s
}

func Start(plugin Plugin, helpProvider externalplugins.ExternalPluginHelpProvider, o CliOptions) {
	lvl, err := logrus.ParseLevel(o.LogLevel)
	if err != nil {
		logrus.WithError(err).Fatal("Could not parse log level.")
	}
	logrus.SetLevel(lvl)
	log := logrus.StandardLogger().WithField("plugin", plugin.GetName())

	if err := secret.Add(o.WebhookSecretPath); err != nil {
		logrus.WithError(err).Fatal("Could not start secret agent.")
	}

	_, err = o.Github.GitClient(o.DryRun)
	if err != nil {
		logrus.WithError(err).Fatal("Could not get git client.")
	}

	mux := http.NewServeMux()
	mux.Handle("/", plugin)
	externalplugins.ServeExternalPluginHelp(mux, log, helpProvider)

	s := http.Server{
		Addr:    ":" + strconv.Itoa(o.Port),
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
