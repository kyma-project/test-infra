package pubsub

import (
	"cloud.google.com/go/pubsub/v2"
	"github.com/kyma-project/test-infra/pkg/logging"
	"google.golang.org/api/option"
)

// ClientConfig holds configuration for pubsub Client.
type ClientConfig struct {
	ProjectID           string
	CredentialsFilePath string
	logger              logging.LoggerInterface
	opts                []option.ClientOption
}

// ClientOption is a client constructor configuration option passing configuration to the constructor.
type ClientOption func(*ClientConfig) error

// Client wraps google pubsub client and provide additional methods.
type Client struct {
	*pubsub.Client
	logger logging.LoggerInterface
}

// Message is the message sent to pubsub system.
type Message struct {
	Message      pubsub.Message `json:"message"`
	Subscription string         `json:"subscription"`
}

// MessagePayload is the pubsub message payload of pubsub message.
type MessagePayload struct {
	Attributes  map[string]string `json:"attributes"`
	Data        []byte            `json:"data"` // This property is base64 encoded
	MessageID   string            `json:"messageId"`
	PublishTime string            `json:"publishTime"`
}

type Rotation struct {
	NextRotationTime string `yaml:"nextRotationTime"`
	RotationPeriod   string `yaml:"rotationPeriod"`
}

// SecretRotateMessage is the Data field of pubsub message payload, published by secret rotation automation.
type SecretRotateMessage struct {
	Name       string              `yaml:"name"`
	CreateTime string              `yaml:"createTime"`
	Labels     map[string]string   `yaml:"labels,omitempty"`
	Topics     []map[string]string `yaml:"topics,omitempty"`
	Etag       string              `yaml:"etag"`
	Rotation   Rotation            `yaml:"rotation,omitempty"`
}
