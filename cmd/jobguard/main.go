package main

import (
	"flag"
	"github.com/kyma-project/test-infra/pkg/jobguard/v2"
	"os"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/prow/prow/config/secret"
	"sigs.k8s.io/prow/prow/flagutil"
)

type options struct {
	debug            bool
	dryRun           bool
	github           flagutil.GitHubOptions
	jobguardOptions  v2.Options
	expContextRegexp string
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options
	fs.BoolVar(&o.debug, "debug", false, "Enable debug logging.")
	fs.BoolVar(&o.dryRun, "dry-run", false, "Enable dry run.")
	fs.StringVar(&o.expContextRegexp, "expected-contexts-regexp", "", "Regular expression with expected contexts.")

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

	var token string
	if o.github.TokenPath == "" {
		logrus.Fatal("Missing github token path")
	}
	token = o.github.TokenPath

	if err := secret.Add(token); err != nil {
		logrus.WithError(err).Fatal("Could not start SecretAgent.")
	}
	logrus.Debugf("%+v", o)

	githubStatus, err := o.github.GitHubClientWithLogFields(o.dryRun, logrus.Fields{"component": "github-status"})
	if err != nil {
		logrus.WithError(err).Fatal("Could not start GitHub client.")
	}

	o.jobguardOptions.PredicateFunc = v2.RegexpPredicate(o.expContextRegexp)
	jobGuardClient := v2.NewClient(githubStatus, o.jobguardOptions)
	if err := jobGuardClient.Run(); err != nil {
		logrus.WithError(err).Fatal("JobGuard caught error.")
	}
	logrus.Infoln("All required checks have successful state.")
}
