package v2

import (
	"flag"
	"fmt"
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
	fs.StringVar(&o.ExpContextRegexp, "expected-contexts-regexp", "", "Regular expression of expected contexts.")
	fs.DurationVar(&o.Timeout, "timeout", DefaultTimeout, "Time after the JobGuard fails.")
	fs.DurationVar(&o.PollInterval, "poll-interval", DefaultPollInterval, "Interval in which JobGuard checks contexts on GitHub.")
	fs.StringVar(&o.Org, "org", "", "Github organisation to check.")
	fs.StringVar(&o.Repo, "repo", "", "GitHub repository to check.")
	fs.StringVar(&o.BaseRef, "base-ref", "", "GitHub base ref to pull statuses from.")
}

func (c Client) Run() error {
	logrus.Info("Building required statuses based on regexp")
	statuses, err := c.BuildStatuses(c.client)
	if err != nil {
		return err
	}
	logrus.Infof("Waiting for statuses to have success state: %v", allStatusesString(statuses))

	timeout := time.After(c.Timeout)
	poller := time.NewTicker(c.PollInterval)
	defer poller.Stop()
loop:
	for {
		select {
		case <-timeout:
			return errors.New("timeout waiting for contexts to be successful")
		case <-poller.C:
			statuses, err = c.Update(c.client, statuses)
			for _, st := range statuses {
				if st.IsPending() {
					logrus.Debugf("Statuses are still pending: %v", allStatusesString(statuses))
					continue loop
				}
				if st.IsFailed() {
					return fmt.Errorf("required context %v is in failed state", st.Context)
				}
			}
		}
		// all checks are successful
		logrus.Debugln(statuses)
		return nil
	}
}

func allStatusesString(st []Status) string {
	var res string
	for i, s := range st {
		res += s.Context
		if i < len(st)-1 {
			res += ", "
		}
	}
	return res
}
