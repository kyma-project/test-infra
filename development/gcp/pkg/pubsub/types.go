package pubsub

import (
	"cloud.google.com/go/pubsub"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

// Client wraps google pubsub client and provide additional methods.
type Client struct {
	*pubsub.Client
}

// This is the message which will be send by pubsub system.
type Message struct {
	Message      MessagePayload `json:"message"`
	Subscription string         `json:"subscription"`
}

// This is the message payload of pubsub message.
type MessagePayload struct {
	Attributes   map[string]string `json:"attributes"`
	Data         []byte            `json:"data"` // This property is base64 encoded
	MessageId    string            `json:"messageId"`
	Message_Id   string            `json:"message_id"`
	PublishTime  string            `json:"publishTime"`
	Publish_time string            `json:"publish_time"`
}

// This is the Data payload of pubsub message payload, published by Prow.
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

// This is the Data payload of pubsub message payload, published by ci-force automation.
// It wraps ProwMessage.
// TODO: consider renaming it to something more generic to use it for other cases
type FailingTestMessage struct {
	ProwMessage
	FirestoreDocumentID   *string  `json:"firestoreDocumentId,omitempty"`
	GithubIssueNumber     *int64   `json:"githubIssueNumber,omitempty"`
	SlackThreadID         *string  `json:"slackThreadId,omitempty"`
	GithubCommitersLogins []string `json:"githubCommiterLogin,omitempty"`
}
