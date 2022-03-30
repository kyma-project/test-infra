package pubsub

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"path"

	"cloud.google.com/go/pubsub"
	"github.com/google/go-github/v40/github"
	"github.com/kyma-project/test-infra/development/logging"
	"google.golang.org/api/option"
)

func (o *ClientConfig) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.ProjectID, "pubsub-project-id", "", "Google cloud pubsub project ID.")
	fs.StringVar(&o.CredentialsFilePath, "pubsub-credentials-files", "/etc/pubsub/credentials.json", "Path to the file with pubsub client credentials.")
}

func (o *ClientConfig) NewClient(ctx context.Context, options ...ClientOption) (*Client, error) {
	var err error
	client := &Client{}
	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option, error: %w", err)
		}
	}

	if o.ProjectID == "" {
		return nil, fmt.Errorf("google pubsub project id was not provided")
	} else if o.CredentialsFilePath == "" {
		return nil, fmt.Errorf("google pubsub client credentials file path was not provided")
	}

	if o.logger != nil {
		client.logger = o.logger
	} else {
		client.logger = logging.NewLogger()
	}

	pubSubClient, err := pubsub.NewClient(ctx, o.ProjectID, o.opts...)
	if err != nil {
		return nil, err
	}
	client.Client = pubSubClient
	return client, nil
}

func (o *ClientConfig) WithLogger(logger logging.LoggerInterface) ClientOption {
	return func(config *ClientConfig) error {
		config.logger = logger
		return nil
	}
}

func (o *ClientConfig) WithGoogleOption(opt option.ClientOption) ClientOption {
	return func(config *ClientConfig) error {
		config.opts = append(config.opts, opt)
		return nil
	}
}

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

func publishPubSubMessage(ctx context.Context, client *pubsub.Client, message interface{}, topicName string, attributes map[string]string) (*string, error) {
	bmessage, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling message to json, error: %w", err)
	}
	topic := client.Topic(topicName)
	result := topic.Publish(ctx, &pubsub.Message{
		// Set json marshaled message as a data payload of pubsub message.
		Data:       bmessage,
		Attributes: attributes,
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
	return publishPubSubMessage(ctx, c.Client, message, topicName, nil)
}

// PublishMessageWithAttributes will send message with attributes to the topicName.
// Message must be anything possible to marshal to json.
// On success publishing it will reply with published message ID.
func (c *Client) PublishMessageWithAttributes(ctx context.Context, message interface{}, topicName string, attributes map[string]string) (*string, error) {
	return publishPubSubMessage(ctx, c.Client, message, topicName, attributes)
}
