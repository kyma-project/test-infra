package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/secret"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/pluginhelp/externalplugins"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	PluginName = "needs-tws"
)

type opts struct {
	port int

	github            prowflagutil.GitHubOptions
	webhookSecretPath string
	logLevel          string
	dryRun            bool
}

func gatherOptions() opts {
	o := opts{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.IntVar(&o.port, "port", 8080, "Plugin port to listen on.")
	fs.BoolVar(&o.dryRun, "dry-run", false, "Run in dry-run mode - no actual changes will be made.")
	fs.StringVar(&o.webhookSecretPath, "hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing GitHub HMAC secret")
	fs.StringVar(&o.logLevel, "log-level", "info", "Set log level.")
	o.github.AddFlags(fs)
	fs.Parse(os.Args[1:])
	return o
}

func main() {
	o := gatherOptions()

	lvl, err := logrus.ParseLevel(o.logLevel)
	if err != nil {
		logrus.WithError(err).Fatal("Could not parse log level.")
	}
	logrus.SetLevel(lvl)
	log := logrus.StandardLogger().WithField("plugin", PluginName)

	if err := secret.Add(o.webhookSecretPath); err != nil {
		logrus.WithError(err).Fatal("Could not start secret agent.")
	}

	ghc, err := o.github.GitHubClient(o.dryRun)
	if err != nil {
		logrus.WithError(err).Fatal("Could not get github client.")
	}
	g, err := o.github.GitClient(o.dryRun)
	if err != nil {
		logrus.WithError(err).Fatal("Could not get git client.")
	}
	gc := git.ClientFactoryFrom(g)
	p := Plugin{
		tokenGenerator: secret.GetTokenGenerator(o.webhookSecretPath),
		ghc:            ghc,
		gc:             gc,
	}

	mux := http.NewServeMux()
	mux.Handle("/", p)
	externalplugins.ServeExternalPluginHelp(mux, log, HelpProvider)

	s := http.Server{
		Addr:    ":" + strconv.Itoa(o.port),
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
