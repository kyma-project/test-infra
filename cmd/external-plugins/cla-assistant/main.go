package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"

	"go.uber.org/zap"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
)

const (
	PluginName = "cla-assistant"
)

var (
	CLACheck = regexp.MustCompile(`(?mi)^/cla-recheck\s*$`)
)

type CLAClient struct {
	address string
	client  http.Client
}

func helpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	ph := &pluginhelp.PluginHelp{
		Description: "Helper plugin that allows quick CLA interactions with CLA Assistant.",
	}
	ph.AddCommand(pluginhelp.Command{
		Usage:       "/cla-recheck",
		Description: "Force CLA check on Pull Request.",
		WhoCanUse:   "Anyone.",
		Examples:    []string{"/cla-recheck"},
	})
	return ph, nil
}

// issueCommentHandler handles all issue_comment webhooks
func (c CLAClient) issueCommentHandler(_ *externalplugin.Plugin, e externalplugin.Event) {
	l := externalplugin.NewLogger().With(
		externalplugin.EventTypeField, e.EventType,
		github.EventGUID, e.EventGUID,
	)
	defer l.Sync()
	l.Debug("Got issue_comment event")
	p := github.IssueCommentEvent{}
	err := json.Unmarshal(e.Payload, &p)
	if err != nil {
		l.Errorw("Could not unmarshal event", "error", err)
	}
	if err := c.handleIssueCommentEvent(l, p); err != nil {
		l.Errorw("Error handling event", "error", err)
	}
}

// handleIssueCommentEvent handles parsed IssueCommentEvent
func (c CLAClient) handleIssueCommentEvent(l *zap.SugaredLogger, ic github.IssueCommentEvent) error {
	if ic.Action != github.IssueCommentActionCreated && !ic.Issue.IsPullRequest() {
		return nil
	}
	if !CLACheck.MatchString(ic.Comment.Body) {
		return nil
	}
	org := ic.Repo.Owner.Login
	repo := ic.Repo.Name
	number := ic.Issue.Number
	l.Debugf("CLA Recheck Request for %s/%s#%d", org, repo, number)
	// HEAD gets the same response as GET
	// If this won't work then change it to GET
	_, err := c.client.Head(fmt.Sprintf("%s/check/%s/%s?pullRequest=%d", c.address, org, repo, number))
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var addr string
	l := externalplugin.NewLogger()
	o := externalplugin.Opts{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.IntVar(&o.Port, "port", 8080, "Plugin port to listen on.")
	fs.BoolVar(&o.DryRun, "dry-run", false, "Run in dry-run mode - no actual changes will be made.")
	fs.StringVar(&o.WebhookSecretPath, "hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing GitHub HMAC secret")
	fs.StringVar(&o.LogLevel, "log-level", "info", "Set log level.")
	fs.StringVar(&addr, "address", "https://cla-assistant.io", "CLA Assistant address.")
	fs.Parse(os.Args[1:])

	c := CLAClient{
		address: addr,
		client:  http.Client{Timeout: time.Minute * 5},
	}

	s := externalplugin.Plugin{
		Name: PluginName,
	}
	s.WithLogger(l)
	s.WithWebhookSecret(o.WebhookSecretPath)
	s.RegisterWebhookHandler("issue_comment", c.issueCommentHandler)
	l.Infof("Start plugin %s", PluginName)
	externalplugin.Start(&s, helpProvider, &o)
}
