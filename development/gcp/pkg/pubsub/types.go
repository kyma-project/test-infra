package pubsub

// This message will be send by pubsub system.
type Message struct {
	Message      MessagePayload `json:"message"`
	Subscription string         `json:"subscription"`
}

// This is the Message payload of pubsub message.
type MessagePayload struct {
	Attributes   map[string]string `json:"attributes"`
	Data         []byte            `json:"data"` // This property is base64 encoded
	MessageId    string            `json:"messageId"`
	Message_Id   string            `json:"message_id"`
	PublishTime  string            `json:"publishTime"`
	Publish_time string            `json:"publish_time"`
}

// This is the Data payload of pubsub message published by Prow.
type ProwMessage struct {
	Project *string `json:"project"`
	Topic   *string `json:"topic"`
	RunID   *string `json:"runid"`
	Status  *string `json:"status"`
	URL     *string `json:"url"`
	GcsPath *string `json:"gcs_path"`
	// TODO: define refs type to force using pointers
	Refs    []map[string]interface{} `json:"refs"`
	JobType *string                  `json:"job_type"`
	JobName *string                  `json:"job_name"`
}

// This is the Data payload of pubsub message published by ci-force automation.
// TODO: consider renaming it to something more generic to use it for other cases
type FailingTestMessage struct {
	ProwMessage
	FirestoreDocumentID *string `json:"firestoreDocumentId,omitempty"`
	GithubIssueNumber   *int64  `json:"githubIssueNumber,omitempty"`
	SlackThreadID       *string `json:"slackThreadId,omitempty"`
}
