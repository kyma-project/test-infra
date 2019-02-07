/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package reporter contains helpers for publishing failed ProwJob statues to Slack channel.
package slack

import (
	"fmt"

	apiSlack "github.com/nlopes/slack"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const (
	// SlackSkipReportLabel annotation
	SlackSkipReportLabel = "prow.k8s.io/slack.skipReportOnFailure"
)

// Client is a reporter client fed to crier controller
type Client struct {
	slackCli     SlackSender
	channel      string
	actOnJobType map[prowapi.ProwJobType]struct{}
}

type SlackSender interface {
	SendMessage(channel string, options ...apiSlack.MsgOption) (string, string, string, error)
}

// NewReporter creates a new Slack reporter
func NewReporter(slackCli SlackSender, channel string, pjTypes []prowapi.ProwJobType) *Client {
	actOnJobType := map[prowapi.ProwJobType]struct{}{}
	for _, typ := range pjTypes {
		actOnJobType[typ] = struct{}{}
	}

	return &Client{
		slackCli:     slackCli,
		channel:      channel,
		actOnJobType: actOnJobType,
	}
}

// GetName returns the name of the reporter
func (c *Client) GetName() string {
	return "slack-reporter"
}

func (c *Client) ShouldReport(pj *prowapi.ProwJob) bool {
	skip := pj.Labels[SlackSkipReportLabel]
	if skip == "true" {
		return false
	}

	_, act := c.actOnJobType[pj.Spec.Type]
	if !act {
		return false
	}

	if !(pj.Status.State == prowapi.ErrorState || pj.Status.State == prowapi.FailureState) {
		return false
	}

	return true
}

// Report takes a ProwJob, and generate a Slack ReportMessage and publish to given channel
func (c *Client) Report(pj *prowapi.ProwJob) error {
	message := c.generateMessageFromPJ(pj)

	c.slackCli.SendMessage(c.channel, apiSlack.MsgOptionText(message, false))

	return nil
}

func (c *Client) generateMessageFromPJ(pj *prowapi.ProwJob) string {
	return fmt.Sprintf("ProwJob %s *FAILED:* State <%s>. Click <%s|here> to see the Job details. ", pj.Spec.Job, pj.Status.State, pj.Status.URL)
	//return &ReportMessage{
	//	Project: pubSubMap[SlackSkipReportLabel],
	//	Topic:   pubSubMap[SlackOverrideChannelLabel],
	//	RunID:   pubSubMap[PubSubRunIDLabel],
	//	Status:  pj.Status.State,
	//	URL:     pj.Status.URL,
	//	GCSPath: strings.Replace(pj.Status.URL, c.config().Plank.JobURLPrefix, GCSPrefix, 1),
	//}
}
