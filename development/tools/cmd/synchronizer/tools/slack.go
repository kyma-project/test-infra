package tools

import (
	"fmt"
	"log"

	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

// Message provide interface for message sent to slack
type Message interface {
	GetTitle() string
	GetValue() string
}

var (
	username      = "Component Alert"
	messageHeader = "*Kyma components are out of date*"
	description   = "Below components are out of date:"
	icon          = ":kyma2:"
	barColor      = "#D96459"
)

// SendMessage sent message to specifuc channel based on slack token
func SendMessage(token, channelName string, msg []Message) error {
	api := slack.New(token)
	channelID, err := findChannelID(api, channelName)
	if err != nil {
		return errors.Wrap(err, "during find channel ID")
	}

	resp, timestamp, err := api.PostMessage(channelID,
		slack.MsgOptionUsername(username),
		slack.MsgOptionText(messageHeader, false),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{IconEmoji: icon, Markdown: true}),
		slack.MsgOptionAttachments(generateMessage(msg)),
	)
	if err != nil {
		return errors.Wrap(err, "during send slack message")
	}

	log.Printf("Slack sent message to channel %q, slack response: %q, slack timestamp: %s", channelName, resp, timestamp)
	return nil
}

func findChannelID(api *slack.Client, name string) (string, error) {
	ch, err := api.GetChannels(false)
	if err != nil {
		return "", errors.Wrap(err, "during fetch channels list")
	}

	for _, slackChannel := range ch {
		if slackChannel.Name == name {
			return slackChannel.ID, nil
		}
	}

	return "", fmt.Errorf("Cannot find channel with name %q on slack", name)
}

func generateMessage(content []Message) slack.Attachment {
	var slackContent []slack.AttachmentField

	for _, part := range content {
		slackContent = append(slackContent, slack.AttachmentField{
			Title: part.GetTitle(),
			Value: part.GetValue(),
		})
	}

	return slack.Attachment{
		Color:   barColor,
		Pretext: description,
		Fields:  slackContent,
	}
}
