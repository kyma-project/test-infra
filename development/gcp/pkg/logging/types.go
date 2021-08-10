package logging

import (
	"cloud.google.com/go/logging"
)

type Payload struct {
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
	Type    string `json:"@type,omitempty"`
}

type Client struct {
	*logging.Client
	LogName string
}

type Logger struct {
	*logging.Logger
	trace   string
	context string
}
