// Package reporter contains helpers for publishing failed ProwJob statues to Slack channel.
// Implementation respects the Crier Reporter interface:
// https://github.com/kubernetes/test-infra/tree/d195f316c99dd376934e6a0ae103b86e6da0db06/prow/cmd/crier#adding-a-new-reporter
package slack

import (
	"fmt"

	apiSlack "github.com/nlopes/slack"
	"github.com/pkg/errors"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const (
	// SlackSkipReportLabel annotation
	SlackSkipReportLabel = "prow.k8s.io/slack.skipReport"

	icon     = ":prow:"
	username = "prow-notifier"
)

// ConfigReporter holds configuration for Slack Reporter
type ConfigReporter struct {
	ActOnProwJobType  []prowapi.ProwJobType  `envconfig:"default=periodic;postsubmit"`
	ActOnProwJobState []prowapi.ProwJobState `envconfig:"default=failure;error"`
	Channel           string
}

// SlackSender sends messages to Slack
type SlackSender interface {
	SendMessage(channel string, options ...apiSlack.MsgOption) (string, string, string, error)
}

// Reporter is a reporter client for slack
type Reporter struct {
	slackCli      SlackSender
	channel       string
	actOnJobType  map[prowapi.ProwJobType]struct{}
	actOnJobState map[prowapi.ProwJobState]struct{}
}

// NewReporter creates a new Slack reporter
func NewReporter(cfg ConfigReporter, slackCli SlackSender) *Reporter {
	actOnJobType := map[prowapi.ProwJobType]struct{}{}
	for _, typ := range cfg.ActOnProwJobType {
		actOnJobType[typ] = struct{}{}
	}

	actOnJobState := map[prowapi.ProwJobState]struct{}{}
	for _, state := range cfg.ActOnProwJobState {
		actOnJobState[state] = struct{}{}
	}

	return &Reporter{
		slackCli:      slackCli,
		channel:       cfg.Channel,
		actOnJobType:  actOnJobType,
		actOnJobState: actOnJobState,
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
	// TODO(mszostok): can be in future replaced with renderer functionality,
	// similar to this one: https://github.com/kyma-project/kyma/blob/release-0.6/tools/stability-checker/internal/notifier/renderer.go
	var (
		header = r.generateHeader(pj)
		body   = r.generateBody(pj)
		footer = r.generateFooter(pj)
	)

	_, _, _, err := r.slackCli.SendMessage(
		r.channel,

		apiSlack.MsgOptionUsername(username),
		apiSlack.MsgOptionPostMessageParameters(apiSlack.PostMessageParameters{IconEmoji: icon, Markdown: true}),

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
		details := []apiSlack.AttachmentField{
			{
				Title: "Repository",
				Value: italic(pj.Spec.Refs.Repo),
				Short: true,
			},
			{
				Title: "Commit Culprit",
				Value: italic(pj.Spec.Refs.BaseSHA),
				Short: false,
			},
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
