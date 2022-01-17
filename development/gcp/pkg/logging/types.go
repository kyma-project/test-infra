package logging

import (
	"cloud.google.com/go/logging"
)

// Payload represent payload which will be send to gcp stackdriver.
type Payload struct {
	// This is the log message.
	Message string `json:"message"`
	// Context provide wider context description for log message.
	Context string `json:"context,omitempty"`
	// Type is set to constant value when log message should reported to google error reporting.
	Type string `json:"@type,omitempty"`
}

// Client wraps google gcp logging client and provides additional methods.
type Client struct {
	*logging.Client
	LogName string
}

// Logger wraps google gcp logging Logger and provides additional methods and fields.
type Logger struct {
	*logging.Logger
	trace   string
	context string
}
