package logging

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"cloud.google.com/go/logging"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

const (
	// ErrorReportingType holds log message payload value for Type expected by google cloud logging to report message as error in Error Reporting.
	ErrorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
	// Default project to store logs.
	LogsProjectID = "sap-kyma-prow"
	// Default google cloud logging log file name for log messages send by prowjobs.
	ProwjobsLogName = "prowjobs"
)

// newClient create kyma implementation of gcp logging client.
// It requires credentials file path to authenticate in GCP.
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

// NewProwjobClient creates kyma implementation google cloud logging client with defaults for using in prowjobs.
// Prow preset with service account credentials for logging to gcp: preset-prowjob-gcp-logging
func NewProwjobClient(ctx context.Context, credentialsFilePath string) (*Client, error) {
	client, err := newClient(ctx, credentialsFilePath, ProwjobsLogName)
	if err != nil {
		return nil, fmt.Errorf("got error while creating prowjob gcp logging client, error: %w", err)
	}
	return client, nil
}

// GetProwjobLabels extract default labels for logging messages in prowjob.
// All data is get from environment variables set by Prow.
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

// NewProwjobLogger creates logger with defaults for logging from prowjobs.
func (c *Client) NewProwjobLogger() *Logger {
	// Get default labels for prowjobs.
	prowjobLabels := GetProwjobLabels()
	l := c.Logger(c.LogName, logging.CommonLabels(prowjobLabels))
	logger := &Logger{Logger: l}
	return logger
}

// WithTrace adds trace value to the logger.
func (l *Logger) WithTrace(trace string) *Logger {
	l.trace = trace
	return l
}

// WithGeneratedTrace generates trace value and adds it to the logger.
func (l *Logger) WithGeneratedTrace() *Logger {
	randomInt := rand.Int()
	// TODO: align trace format for all components
	l.trace = fmt.Sprintf("trace/%d", randomInt)
	return l
}

// WithContext sets context value for logger.
// Because there can be multiple contexts in an application it crates Logger copy before setting the context.
func (l Logger) WithContext(entryContext string) Logger {
	l.context = entryContext
	return l
}

// LogError log message of Error severity with defaults. Error message is send to standard logger as well
// It will set log message payload Type to value required by gcp stackdriver to report messages as error in Error Reporting.
func (l *Logger) LogError(message string) {
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
	log.Error(message)
}

// LogInfo log message of Info severity wit defaults. Info message is send to standard logger as well.
func (l *Logger) LogInfo(message string) {
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
	log.Info(message)
}
