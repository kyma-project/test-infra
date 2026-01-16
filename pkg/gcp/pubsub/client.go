package pubsub

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"path"

	"cloud.google.com/go/pubsub/v2"
	"github.com/google/go-github/v80/github"
	"github.com/kyma-project/test-infra/pkg/logging"
	"google.golang.org/api/option"
)

// AddFlags add pubsub client flags to provided flagset.
// Flag set can be parsed along with other flags.
func (o *ClientConfig) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.ProjectID, "pubsub-project-id", "", "Google cloud pubsub project ID.")
	fs.StringVar(&o.CredentialsFilePath, "pubsub-credentials-files", "/etc/pubsub/credentials.json", "Path to the file with pubsub client credentials.")
}

// NewClient is a pubsub client wrapper construction function. Client can be configured by providing ClientOptions to the constructor.
// Constructor provide console logger as default logger.
func (o *ClientConfig) NewClient(ctx context.Context, options ...ClientOption) (*Client, error) {
	var err error
	client := &Client{}

	// Go through functional options.
	for _, opt := range options {
		err := opt(o)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option, error: %w", err)
		}
	}

	// Check if mandatory option is provided
	if o.ProjectID == "" {
		return nil, fmt.Errorf("google pubsub project id was not provided")
	} else if o.CredentialsFilePath == "" {
		return nil, fmt.Errorf("google pubsub client credentials file path was not provided")
	}

	// Create default logger if not provided.
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

// WithLogger is constructor function configuration option providing logger instance to use with client.
func (o *ClientConfig) WithLogger(logger logging.LoggerInterface) ClientOption {
	return func(config *ClientConfig) error {
		config.logger = logger
		return nil
	}
}

// WithGoogleOption is a constructor function configuration option providing google ClientOption to pass to google pubsub client constructor.
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

// GetJobID will extract prowjob  ID from prowjob URL. Prowjob ID is a last element of prowjob URL.
func GetJobID(jobURL *string) (*string, error) {
	parsedJobURL, err := url.Parse(*jobURL)
	if err != nil {
		return nil, fmt.Errorf("failed parse test URL, error: %w", err)
	}
	jobID := path.Base(parsedJobURL.Path)
	return github.Ptr(jobID), nil
}

// publishPubSubMessage construct pubsub message and publish to pubsub topic.
// Function message argument will be used as pubsub message payload.
// Function topicName argument will be used as a topic to publish message too.
func (c *Client) publishPubSubMessage(ctx context.Context, message interface{}, topicName string, attributes map[string]string) (*string, error) {
	bmessage, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling message to json, error: %w", err)
	}
	topic := c.Publisher(topicName)
	result := topic.Publish(ctx, &pubsub.Message{
		// Set json marshaled message as a data payload of pubsub message.
		Data:       bmessage,
		Attributes: attributes,
	})
	publishedID, err := result.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed publishing to topic %s, error: %w", topicName, err)
	}
	return github.Ptr(publishedID), nil
}

// PublishMessage will send message to the topicName. Message must be anything possible to marshal to json.
// On success publishing it will reply with published message ID.
func (c *Client) PublishMessage(ctx context.Context, message interface{}, topicName string) (*string, error) {
	return c.publishPubSubMessage(ctx, message, topicName, nil)
}

// PublishMessageWithAttributes will send message with attributes to the topicName.
// Message must be anything possible to marshal to json.
// On success publishing it will reply with published message ID.
func (c *Client) PublishMessageWithAttributes(ctx context.Context, message interface{}, topicName string, attributes map[string]string) (*string, error) {
	return c.publishPubSubMessage(ctx, message, topicName, attributes)
}
