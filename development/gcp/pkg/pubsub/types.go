package pubsub

import (
	"cloud.google.com/go/pubsub"
	"github.com/kyma-project/test-infra/development/logging"
	"google.golang.org/api/option"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
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
	Message      MessagePayload `json:"message"`
	Subscription string         `json:"subscription"`
}

// MessagePayload is the pubsub message payload of pubsub message.
type MessagePayload struct {
	Attributes  map[string]string `json:"attributes"`
	Data        []byte            `json:"data"` // This property is base64 encoded
	MessageID   string            `json:"messageId"`
	PublishTime string            `json:"publishTime"`
}

// ProwMessage is the Data field of pubsub message payload, published by Prow.
type ProwMessage struct {
	Project *string `json:"project"`
	Topic   *string `json:"topic"`
	RunID   *string `json:"runid"`
	Status  *string `json:"status"`
	URL     *string `json:"url"`
	GcsPath *string `json:"gcs_path"`
	// TODO: define refs type to force using pointers
	Refs    []prowapi.Refs `json:"refs"`
	JobType *string        `json:"job_type"`
	JobName *string        `json:"job_name"`
}

// FailingTestMessage is the Data field of pubsub message payload, published by ci-force automation.
// It wraps ProwMessage.
// TODO: consider renaming it to something more generic to use it for other cases
type FailingTestMessage struct {
	ProwMessage
	FirestoreDocumentID   *string  `json:"firestoreDocumentId,omitempty"`
	GithubIssueNumber     *int64   `json:"githubIssueNumber,omitempty"`
	GithubIssueRepo       *string  `json:"githubIssueRepo,omitempty"`
	GithubIssueOrg        *string  `json:"githubIssueOrg,omitempty"`
	GithubIssueURL        *string  `json:"githubIssueUrl,omitempty"`
	SlackThreadID         *string  `json:"slackThreadId,omitempty"`
	GithubCommitersLogins []string `json:"githubCommitersLogins,omitempty"`
	CommitersSlackLogins  []string `json:"slackCommitersLogins,omitempty"`
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
