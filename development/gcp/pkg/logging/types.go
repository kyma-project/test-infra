package logging

import (
	"cloud.google.com/go/logging"
)

type loggingPayload struct {
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
	Type    string `json:"@type,omitempty"`
}

type PeriodicProwjobLabels struct {
	JobName   string
	JobType   string
	BuildID   string
	ProwjobID string
}

type PostsubmitProwjobLabels struct {
	PeriodicProwjobLabels
	RepoName  string
	CommitSHA string
}

type PresubmitProwjobLabels struct {
	PostsubmitProwjobLabels
	PrNumber string
	PrSHA    string
}

type Client struct {
	*logging.Client
	LogName string
}
