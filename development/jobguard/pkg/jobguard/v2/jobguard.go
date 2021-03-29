package v2

import (
	"flag"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/github"
	"time"
)

const (
	DefaultTimeout      = 15 * time.Minute
	DefaultPollInterval = 1 * time.Minute
)

type Options struct {
	Timeout      time.Duration
	PollInterval time.Duration
	StatusOptions
}

type Client struct {
	client github.Client
	Options
}

func NewClient(client github.Client, opts Options) *Client {
	c := new(Client)
	c.client = client
	c.Options = opts
	return c
}

func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.FailOnNoContexts, "fail-on-no-contexts", false, "Fail if regexp does not match to any of the GitHub contexts.")
	fs.DurationVar(&o.Timeout, "timeout", DefaultTimeout, "Time after the JobGuard fails.")
	fs.DurationVar(&o.PollInterval, "poll-interval", DefaultPollInterval, "Interval in which JobGuard checks contexts on GitHub.")
	fs.StringVar(&o.Org, "org", "", "Github organisation to check.")
	fs.StringVar(&o.Repo, "repo", "", "GitHub repository to check.")
	fs.StringVar(&o.BaseRef, "base-ref", "", "GitHub base ref to pull statuses from.")
}

func (c Client) Run() error {
	logrus.Info("Building required statuses based on regexp")
	statuses, err := c.FetchRequiredStatuses(c.client, c.PredicateFunc)
	if err != nil {
		return err
	}
	logrus.Infof("Waiting for statuses to have success state: %v", statuses)

	timeout := time.After(c.Timeout)
	poller := time.NewTicker(c.PollInterval)
	defer poller.Stop()

	for {
		statuses, err := c.Update(c.client, statuses)
		if err != nil {
			return err
		}
		switch statuses.CombinedStatus() {
		case StatusPending:
			logrus.Infof("Some statuses are still in pending state.")
			break
		case StatusFailure:
			return errors.New("statuses are in failed state")
		case StatusSuccess:
			return nil
		}
		select {
		case <-timeout:
			return errors.New("timeout waiting for contexts to be successful")
		case <-poller.C:
		}
	}
}
