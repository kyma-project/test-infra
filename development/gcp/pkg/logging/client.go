package logging

import (
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"math/rand"
	"os"
	"time"
)

const (
	ErrorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
	LogsProjectID      = "sap-kyma-prow"
	ProwjobsLogName    = "prowjobs"
)

func newClient(ctx context.Context, credentialsFilePath, logName string) (*Client, error) {
	client := &Client{
		LogName: logName,
	}
	c, err := logging.NewClient(ctx, LogsProjectID, option.WithCredentialsFile(credentialsFilePath))
	if err != nil {
		return nil, fmt.Errorf("got error while creating google cloud logging client, error: %w", err)
	}
	client.Client = c
	return client, nil
}

// TODO: consider moving this to development/prow/logging.go
func NewProwjobClient(ctx context.Context, credentialsFilePath string) (*Client, error) {
	client, err := newClient(ctx, credentialsFilePath, ProwjobsLogName)
	if err != nil {
		return nil, fmt.Errorf("got error while creating prowjob gcp logging client, error: %w", err)
	}
	return client, nil
}

func GetProwjobLabels() map[string]string {
	jobType := os.Getenv("JOB_TYPE")
	labels := map[string]string{
		"jobName":   os.Getenv("JOB_NAME"),
		"jobType":   os.Getenv("JOB_TYPE"),
		"buildID":   os.Getenv("BUILD_ID"),
		"prowjobID": os.Getenv("PROW_JOB_ID"),
	}
	if jobType == "presubmit" || jobType == "postsubmit" {
		labels["repoName"] = os.Getenv("REPO_NAME")
		labels["commitSHA"] = os.Getenv("PULL_BASE_SHA")
	}
	if jobType == "presubmit" {
		labels["prNumber"] = os.Getenv("PULL_NUMBER")
		labels["prSHA"] = os.Getenv("PULL_PULL_SHA")
	}
	return labels
}

func (c *Client) NewProwjobLogger() *Logger {
	prowjobLabels := GetProwjobLabels()
	l := c.Logger(c.LogName, logging.CommonLabels(prowjobLabels))
	logger := &Logger{Logger: l}
	return logger
}

func (l *Logger) WithTrace(trace string) *Logger {
	l.trace = trace
	return l
}

func (l *Logger) WithGeneratedTrace() *Logger {
	randomInt := rand.Int()
	// TODO: align trace format for all components
	l.trace = fmt.Sprintf("trace/%d", randomInt)
	return l
}

func (l Logger) WithContext(entryContext string) Logger {
	l.context = entryContext
	return l
}

func (l Logger) LogError(message string) {
	entry := logging.Entry{
		Timestamp: time.Now(),
		Severity:  logging.Error,
	}
	payload := Payload{
		Message: message,
		Type:    ErrorReportingType,
	}
	if l.context != "" {
		payload.Context = l.context
	}
	entry.Payload = payload
	if l.trace != "" {
		entry.Trace = l.trace
	}
	l.Log(entry)
	log.Errorf(message)
}

func (l Logger) LogInfo(message string) {
	entry := logging.Entry{
		Timestamp: time.Now(),
		Severity:  logging.Info,
	}
	payload := Payload{
		Message: message,
	}
	if l.context != "" {
		payload.Context = l.context
	}
	entry.Payload = payload
	if l.trace != "" {
		entry.Trace = l.trace
	}
	l.Log(entry)
	log.Errorf(message)
}
