package slack

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	logf "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	octopusTypes "github.com/kyma-project/test-infra/development/test-log-collector/pkg/resources/clustertestsuite/types"
)

type Attributes struct {
	Name             string
	Status           string
	ClusterTestSuite string
	CompletionTime   string
	Platform         string
}

type Message struct {
	Data        string
	Attributes  Attributes
	ChannelName string
	ChannelID   string
	PodName     string
}

type Client struct {
	client *slack.Client
}

func New(client *slack.Client) *Client {
	return &Client{
		client: client,
	}
}

func (s Client) parentMessageTimestamp(hist slack.GetConversationHistoryResponse, parentMsg string) (string, bool) {
	for _, msg := range hist.Messages {
		if msg.Text == parentMsg {
			return msg.Timestamp, true
		}
	}
	return "", false
}

func (s Client) createParentMessage(channelID, parentMessage string) error {
	hist, err := s.client.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit: 100, // should be more than enough
	})

	if err != nil {
		return errors.Wrapf(err, "while getting channel historical messages by id: %s", channelID)
	}

	_, exists := s.parentMessageTimestamp(*hist, parentMessage)
	if exists {
		logf.Info("parent message already exists")
		return nil
	}

	logf.Info("creating parent message")

	_, _, err = s.client.PostMessage(channelID, slack.MsgOptionText(parentMessage, false))
	if err != nil {
		return errors.Wrap(err, "while creating slack thread")
	}
	return nil
}

func (s Client) UploadLogFiles(messages []Message, ctsName, completionTime, platform string) error {
	failedTestNames := getFailedTestNames(messages)

	for channelID, messageSlice := range s.groupMessagesByChannelID(messages) {

		parentMsg := fmt.Sprintf("ClusterTestSuite `%s`; completionTime `%s`; platform `%s`; other failed test names: `%s`", ctsName, completionTime, platform, strings.Join(failedTestNames, ", "))

		if err := s.createParentMessage(channelID, parentMsg); err != nil {
			return errors.Wrapf(err, "while creating parent slack message in channel %s", messageSlice[0].ChannelName)
		}

		// get channel history has to be called here *again*, otherwise slack api acts crazy
		hist, err := s.client.GetConversationHistory(&slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     100, // should be more than enough
		})
		if err != nil {
			return errors.Wrapf(err, "while getting %s channel historical messages", messageSlice[0].ChannelName)
		}

		parentMsgTimestamp, _ := s.parentMessageTimestamp(*hist, parentMsg)

		for _, msg := range messageSlice {
			if err := s.UploadLogFile(msg, parentMsgTimestamp); err != nil {
				return errors.Wrapf(err, "while uploading logs for %s test case", msg.Attributes.Name)
			}
		}
	}

	return nil
}

func getFailedTestNames(messages []Message) []string {
	var names []string

	for _, msg := range messages {
		if msg.Attributes.Status == string(octopusTypes.SuiteFailed) {
			names = append(names, msg.Attributes.Name)
		}
	}

	return names
}

func (s Client) groupMessagesByChannelID(messages []Message) map[string][]Message {
	mp := make(map[string][]Message, 0)
	for _, msg := range messages {
		mp[msg.ChannelID] = append(mp[msg.ChannelID], msg)
	}
	return mp
}

func (s Client) UploadLogFile(msg Message, parentMsgTimestamp string) error {
	logf.Info("uploading log file")
	_, err := s.client.UploadFile(slack.FileUploadParameters{
		Content:        msg.Data,
		Filename:       "logs.txt",
		Title:          "Test logs",
		InitialComment: fmt.Sprintf("Test: `%s`, status: `%s`, pod name: `%s`", msg.Attributes.Name, msg.Attributes.Status, msg.PodName),
		Channels: []string{
			msg.ChannelID,
		},
		ThreadTimestamp: parentMsgTimestamp,
	})
	if err != nil {
		return err
	}

	return nil
}
