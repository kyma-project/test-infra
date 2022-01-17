package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v40/github"
	"net/url"
	"path"
)

// NewClient create kyma implementation of pubsub Client.
// It wraps google pubsub client.
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Client{Client: pubSubClient}, nil
}

// GetJobId will extract prowjob  ID from prowjob URL. Prowjob ID is a last element of prowjob URL.
func GetJobId(jobUrl *string) (*string, error) {
	jobURL, err := url.Parse(*jobUrl)
	if err != nil {
		return nil, fmt.Errorf("failed parse test URL, error: %w", err)
	}
	jobID := path.Base(jobURL.Path)
	return github.String(jobID), nil
}

// TODO: remove calls to this function. Calls should be replaced with calls to client method PublishMessage.
func PublishPubSubMessage(ctx context.Context, client *pubsub.Client, message interface{}, topicName string) (*string, error) {
	bmessage, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling message to json, error: %w", err)
	}
	topic := client.Topic(topicName)
	result := topic.Publish(ctx, &pubsub.Message{
		// Set json marshaled message as a data payload of pubsub message.
		Data: bmessage,
	})
	publishedID, err := result.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed publishing to topic %s, error: %w", topicName, err)
	}
	return github.String(publishedID), nil
}

// PublishMessage will send message to the topicName. Message must be anything possible to marshal to json.
// On success publishing it will reply with published message ID.
func (c *Client) PublishMessage(ctx context.Context, message interface{}, topicName string) (*string, error) {
	return PublishPubSubMessage(ctx, c.Client, message, topicName)
}
