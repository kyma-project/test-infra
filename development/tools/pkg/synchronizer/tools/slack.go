package tools

import (
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
	messageHeader = "*The component's version diverges from the version of the latest commit.*"
	description   = "These components versions are out-of-date:"
	icon          = ":kyma2:"
	barColor      = "#D96459"
)

// SendMessage sent message to specifuc channel based on slack token
func SendMessage(token, channelName string, msg []Message) error {
	api := slack.New(token)

	resp, timestamp, err := api.PostMessage(channelName,
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
