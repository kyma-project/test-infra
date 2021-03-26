package main

import (
	"flag"
	jobguard "github.com/kyma-project/test-infra/development/jobguard/pkg/jobguard/v2"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config/secret"
	"k8s.io/test-infra/prow/flagutil"
	"os"
)

type options struct {
	debug           bool
	dryRun          bool
	github          flagutil.GitHubOptions
	jobguardOptions jobguard.Options
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options
	fs.BoolVar(&o.debug, "debug", false, "Enable debug logging.")
	fs.BoolVar(&o.dryRun, "dry-run", false, "Enable dry run.")

	o.jobguardOptions.AddFlags(fs)
	o.github.AddFlags(fs)

	fs.Parse(args)
	return o
}

func main() {
	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if o.debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logger := logrus.WithField("tool", "jobguard")
	var secretAgent = &secret.Agent{}
	var token string
	if o.github.TokenPath == "" {
		logger.Fatal("Missing github token path")
	}
	token = o.github.TokenPath

	if err := secretAgent.Start([]string{token}); err != nil {
		logger.WithError(err).Fatal("Could not start SecretAgent.")
	}
	logger.Debugf("%+v", o)

	githubStatus, err := o.github.GitHubClientWithLogFields(secretAgent, o.dryRun, logrus.Fields{"component": "github-status"})
	if err != nil {
		logger.WithError(err).Fatal("Could not start GitHub client.")
	}

	jobGuardClient := jobguard.NewClient(githubStatus, o.jobguardOptions)
	if err := jobGuardClient.Run(); err != nil {
		logger.WithError(err).Fatal("JobGuard caught error.")
	}
	logger.Infoln("All required checks have successful state.")
}
