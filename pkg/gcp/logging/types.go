package logging

import (
	"cloud.google.com/go/logging"
)

// Config hnews configuration for GCP logging client.
// It can be passed to the client constructor with client constructor configuration option.
type Config struct {
	AppName             string `envconfig:"APP_NAME"` // PubSub Connector application name as set in Compass.
	LogName             string `envconfig:"LOG_NAME"` // Google cloud logging log name.
	Component           string `envconfig:"COMPONENT"`
	ProjectID           string `envconfig:"LOGGING_GCP_PROJECT_ID"`
	credentialsFilePath string `envconfig:"LOGGING_SA_CREDENTIALS_FILE_PATH"`
	commonLabels        map[string]string
	trace               string
	context             string
}

// ClientOption is a client constructor configuration option.
type ClientOption func(*Config) error

// LoggerOption is a logger constructor configuration option.
type LoggerOption func(*Config) error

// Payload represent payload send to gcp stackdriver.
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
}

// Logger wraps google gcp logging Logger and provides additional methods and fields.
type Logger struct {
	*logging.Logger
	trace   string
	context string
}
# (2025-03-04)