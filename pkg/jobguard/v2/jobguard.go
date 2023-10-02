package v2

import (
	"flag"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/github"
)

const (
	DefaultTimeout      = 15 * time.Minute
	DefaultPollInterval = 1 * time.Minute
)

// Options holds configuration to the JobGuard client.
type Options struct {
	Timeout      time.Duration
	PollInterval time.Duration
	StatusOptions
}

// Client represents JobGuard instance
type Client struct {
	client github.Client
	Options
}

// NewClient returns new Client instance with set of Options
func NewClient(client github.Client, opts Options) *Client {
	c := new(Client)
	c.client = client
	c.Options = opts
	return c
}

// AddFlags configures basic FlagSet for the binary
func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.FailOnNoContexts, "fail-on-no-contexts", false, "Fail if regexp does not match to any of the GitHub contexts.")
	fs.DurationVar(&o.Timeout, "timeout", DefaultTimeout, "Time after the JobGuard fails.")
	fs.DurationVar(&o.PollInterval, "poll-interval", DefaultPollInterval, "Interval in which JobGuard checks contexts on GitHub.")
	fs.StringVar(&o.Org, "org", "", "GitHub organisation to check.")
	fs.StringVar(&o.Repo, "repo", "", "GitHub repository to check.")
	fs.StringVar(&o.BaseRef, "base-ref", "", "GitHub base ref to pull statuses from.")
}

// Run fetches the statuses by a defined StatusPredicate, then updates their status periodically
// until all statuses are in "success" or "failure" state. If the Status is in failed state then returns proper error.
func (c Client) Run() error {
	<-time.After(30 * time.Second)
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
		logrus.Info("Updating statuses...")
		statuses, err := c.Update(c.client, statuses)
		if err != nil {
			return err
		}
		switch statuses.CombinedStatus() {
		case StatusPending:
			logrus.Debugf("Some statuses are still in pending state: %v", statuses.PendingList())
		case StatusFailure:
			return fmt.Errorf("statuses are in failed state: %v", statuses.FailedList())
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
