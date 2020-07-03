package slack

import (
	"fmt"

	"github.com/pkg/errors"
	logf "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
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

type CLient struct {
	client *slack.Client
}

func New(client *slack.Client) *CLient {
	return &CLient{
		client: client,
	}
}

func (s CLient) parentMessageTimestamp(hist slack.History, parentMsg string) (string, bool) {
	for _, msg := range hist.Messages {
		if msg.Text == parentMsg {
			return msg.Timestamp, true
		}
	}
	return "", false
}

func (s CLient) createParentMessage(ctsName, channelID, completionTime, platform string) error {
	hist, err := s.client.GetChannelHistory(channelID, slack.HistoryParameters{
		Count: 100,
	})
	if err != nil {
		return errors.Wrapf(err, "while getting channel historical messages by id: %s", channelID)
	}

	parentMessage := fmt.Sprintf("ClusterTestSuite %s, completionTime %s, platform %s", ctsName, completionTime, platform)

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

func (s CLient) UploadLogFiles(messages []Message, ctsName, completionTime, platform string) error {
	for channelID, messageSlice := range s.groupMessagesByChannelID(messages) {
		if err := s.createParentMessage(ctsName, channelID, completionTime, platform); err != nil {
			return errors.Wrapf(err, "while creating parent slack message in channel %s", messageSlice[0].ChannelName)
		}

		// get channel history has to be called here *again*, otherwise slack api acts crazy
		hist, err := s.client.GetChannelHistory(channelID, slack.HistoryParameters{
			Count: 100, // it should be more than enough
		})
		if err != nil {
			return errors.Wrapf(err, "while getting %s channel historical messages", messageSlice[0].ChannelName)
		}

		parentMessage := fmt.Sprintf("ClusterTestSuite %s, completionTime %s, platform %s", ctsName, completionTime, platform)

		parentMsgTimestamp, _ := s.parentMessageTimestamp(*hist, parentMessage)

		for _, msg := range messageSlice {
			if err := s.UploadLogFile(msg, parentMsgTimestamp); err != nil {
				return errors.Wrapf(err, "while uploading logs for %s test case", msg.Attributes.Name)
			}
		}
	}

	return nil
}

func (s CLient) groupMessagesByChannelID(messages []Message) map[string][]Message {
	mp := make(map[string][]Message, 0)
	for _, msg := range messages {
		mp[msg.ChannelID] = append(mp[msg.ChannelID], msg)
	}
	return mp
}

func (s CLient) UploadLogFile(msg Message, parentMsgTimestamp string) error {
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
