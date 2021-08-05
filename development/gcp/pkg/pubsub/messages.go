package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v36/github"
	"net/url"
	"path"
)

func GetJobId(jobUrl *string) (*string, error) {
	jobURL, err := url.Parse(*jobUrl)
	if err != nil {
		return nil, fmt.Errorf("failed parse test URL, error: %w", err)
	}
	jobID := path.Base(jobURL.Path)
	return github.String(jobID), nil
}

func PublishPubSubMessage(ctx context.Context, client *pubsub.Client, message interface{}, topicName string) (*string, error) {
	bmessage, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling message to json, error: %w", err)
	}
	topic := client.Topic(topicName)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: bmessage,
	})
	publishedID, err := result.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed publishing to topic %s, error: %w", topicName, err)
	}
	return github.String(publishedID), nil
}
