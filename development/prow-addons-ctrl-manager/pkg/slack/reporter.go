// Package reporter contains helpers for publishing failed ProwJob statues to Slack channel.
// Implementation respects the Crier Reporter interface:
// https://github.com/kubernetes/test-infra/tree/d195f316c99dd376934e6a0ae103b86e6da0db06/prow/cmd/crier#adding-a-new-reporter
package slack

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/go-github/github"
	apiSlack "github.com/nlopes/slack"
	"github.com/pkg/errors"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	// SlackSkipReportLabel annotation
	SlackSkipReportLabel = "prow.kyma-project.io/slack.skipReport"
)

// ConfigReporter holds configuration for Slack Reporter
type ConfigReporter struct {
	ActOnProwJobType  []prowapi.ProwJobType  `envconfig:"default=periodic;postsubmit"`
	ActOnProwJobState []prowapi.ProwJobState `envconfig:"default=failure;error"`
	Channel           string
	UserIconEmoji     string `envconfig:"default=:prow:"`
	Username          string `envconfig:"default=prow-notifier"`
}

// SlackSender sends messages to Slack
type SlackSender interface {
	SendMessage(channel string, options ...apiSlack.MsgOption) (string, string, string, error)
}

type GithubCommitFetcher interface {
	GetCommit(ctx context.Context, owner string, repo string, sha string) (*github.Commit, *github.Response, error)
}

// Reporter is a reporter client for slack
type Reporter struct {
	channel  string
	icon     string
	username string

	slackCli      SlackSender
	commitFetcher GithubCommitFetcher
	actOnJobType  map[prowapi.ProwJobType]struct{}
	actOnJobState map[prowapi.ProwJobState]struct{}
	log           logr.Logger
}

// NewReporter creates a new Slack reporter
func NewReporter(cfg ConfigReporter, slackCli SlackSender, commitFetcher GithubCommitFetcher) *Reporter {
	actOnJobType := map[prowapi.ProwJobType]struct{}{}
	for _, typ := range cfg.ActOnProwJobType {
		actOnJobType[typ] = struct{}{}
	}

	actOnJobState := map[prowapi.ProwJobState]struct{}{}
	for _, state := range cfg.ActOnProwJobState {
		actOnJobState[state] = struct{}{}
	}

	return &Reporter{
		channel:       cfg.Channel,
		slackCli:      slackCli,
		commitFetcher: commitFetcher,
		actOnJobType:  actOnJobType,
		actOnJobState: actOnJobState,
		icon:          cfg.UserIconEmoji,
		username:      cfg.Username,
		log:           log.Log.WithName("reporter:slack"),
	}
}

// GetName returns the name of the reporter
func (r *Reporter) GetName() string {
	return "slack-reporter"
}

// ShouldReport checks if should react on given ProwJob
func (r *Reporter) ShouldReport(pj *prowapi.ProwJob) bool {
	skip := pj.Labels[SlackSkipReportLabel]
	if skip == "true" {
		return false
	}

	if _, found := r.actOnJobType[pj.Spec.Type]; !found {
		return false
	}

	if _, found := r.actOnJobState[pj.Status.State]; !found {
		return false
	}

	return true
}

// Report takes a ProwJob, and generate a Slack ReportMessage and publish to given channel
func (r *Reporter) Report(pj *prowapi.ProwJob) error {
	// TODO(mszostok): in future  can be replaced with renderer functionality,
	// similar to this one: https://github.com/kyma-project/kyma/blob/release-0.6/tools/stability-checker/internal/notifier/renderer.go
	var (
		header = r.generateHeader(pj)
		body   = r.generateBody(pj)
		footer = r.generateFooter(pj)
	)

	_, _, _, err := r.slackCli.SendMessage(
		r.channel,

		apiSlack.MsgOptionUsername(r.username),
		apiSlack.MsgOptionPostMessageParameters(apiSlack.PostMessageParameters{IconEmoji: r.icon, Markdown: true}),

		apiSlack.MsgOptionText(header, false),
		apiSlack.MsgOptionAttachments(body, footer),
	)
	if err != nil {
		return errors.Wrap(err, "while sending slack message")
	}

	return nil
}

func (r *Reporter) generateHeader(pj *prowapi.ProwJob) string {
	return fmt.Sprintf("*<%s|Prow Job> Alert*", pj.Status.URL)
}

func (r *Reporter) generateBody(pj *prowapi.ProwJob) apiSlack.Attachment {
	var (
		blue   = "#007FFF"
		italic = func(s interface{}) string { return fmt.Sprintf("_%s_", s) }
	)

	body := apiSlack.Attachment{
		Color: blue,
		Fields: []apiSlack.AttachmentField{
			{
				Title: "Name",
				Value: italic(pj.Spec.Job),
				Short: true,
			},
			{
				Title: "State",
				Value: italic(pj.Status.State),
				Short: true,
			},
			{
				Title: "Type",
				Value: italic(pj.Spec.Type),
				Short: true,
			},
		},
	}

	if pj.Spec.Refs != nil {
		org, repo, sha := pj.Spec.Refs.Org, pj.Spec.Refs.Repo, pj.Spec.Refs.BaseSHA
		details := []apiSlack.AttachmentField{
			{
				Title: "Repository",
				Value: italic(repo),
				Short: true,
			},
			{
				Title: "Commit",
				Value: sha[:7],
				Short: true,
			},
		}

		commit, _, err := r.commitFetcher.GetCommit(context.TODO(), org, repo, sha)
		if err != nil {
			r.log.Error(err, "Cannot fetch git commit details", "org", org, "repo", repo, "sha", sha)
		} else {
			details = append(details, apiSlack.AttachmentField{
				Title: "Author",
				Value: italic(commit.Author.GetName()),
				Short: true,
			})
		}

		body.Fields = append(body.Fields, details...)
	}

	return body
}

func (r *Reporter) generateFooter(pj *prowapi.ProwJob) apiSlack.Attachment {
	footer := apiSlack.Attachment{
		Fields: []apiSlack.AttachmentField{
			{
				Value: fmt.Sprintf("See the Job details *<%s|here>.*", pj.Status.URL),
				Short: false,
			},
		},
	}

	return footer
}
