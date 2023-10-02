package logging

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"cloud.google.com/go/logging"
	"google.golang.org/api/option"
)

const (
	// ErrorReportingType holds log message payload value for Type expected by google cloud logging to report message as error in Error Reporting.
	ErrorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
	// ProwLogsProjectID is a default project to store logs.
	ProwLogsProjectID   = "sap-kyma-prow"
	CredentialsFilePath = "/etc/gcpLoggingServiceAccountKey/key"
	// ProwjobsLogName is a default Google Cloud Logging log filename for messages sent by prowjobs.
	ProwjobsLogName = "prowjobs"
	// RepoOwnersServiceLogName is a default Google Cloud Logging log filename for messages sent by repoowners service.
	RepoOwnersServiceLogName = "repoowners"
)

// newClient creates google logging client.
func newClient(ctx context.Context, credentialsFilePath, projectID string) (*logging.Client, error) {
	c, err := logging.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsFilePath))
	if err != nil {
		return nil, fmt.Errorf("got error while creating google cloud logging client, error: %w", err)
	}
	return c, nil
}

// NewClient is a constructor function creating general purpose kyma wrapper of gcp logging client.
// It requires credentials file path to authenticate in GCP.
// A constructor can be configured by providing ClientOptions.
// Constructor provides default logs Google project ID and path to credentials file.
func NewClient(ctx context.Context, options ...ClientOption) (*Client, error) {
	conf := &Config{
		AppName:             "",
		LogName:             "",
		Component:           "",
		ProjectID:           ProwLogsProjectID,
		credentialsFilePath: CredentialsFilePath,
	}

	// Go through provided client options to configure constructor.
	for _, opt := range options {
		err := opt(conf)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	// Get Google client instance.
	c, err := newClient(ctx, conf.ProjectID, conf.credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("got error while creating google cloud logging client, error: %w", err)
	}
	client := &Client{}
	client.Client = c
	return client, nil
}

// WithProjectID is a client constructor configuration option passing GCP project ID to send log messages to.
func WithProjectID(projectID string) ClientOption {
	return func(conf *Config) error {
		conf.ProjectID = projectID
		return nil
	}
}

// WithCredentialsFilePath is a client constructor configuration option passing path to GCP credentials file.
func WithCredentialsFilePath(credentialsFilePath string) ClientOption {
	return func(conf *Config) error {
		conf.credentialsFilePath = credentialsFilePath
		return nil
	}
}

// NewProwjobClient creates kyma wrapper of google cloud logging client with defaults for using in prowjobs.
// TODO: preset-prowjob-gcp-logging should be changed to preset-gcp-logging
// Prow preset with service account credentials for logging to gcp: preset-prowjob-gcp-logging
// It provides default Google logging log name for logging from prowjobs.
func NewProwjobClient(ctx context.Context, credentialsFilePath, gcpproject string) (*Client, error) {
	c, err := newClient(ctx, credentialsFilePath, gcpproject)
	if err != nil {
		return nil, fmt.Errorf("got error while creating prowjob gcp logging client, error: %w", err)
	}
	client := &Client{}
	client.Client = c
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

// NewLogger is a logger constructor function. Logger is created from Google logging client.
// It creates general purpose Google logging logger.
// It provides default Google logging project ID and path to credentials file.
func (c *Client) NewLogger(options ...LoggerOption) (*Logger, error) {
	conf := &Config{
		AppName:             "",
		LogName:             "",
		Component:           "",
		ProjectID:           ProwLogsProjectID,
		credentialsFilePath: CredentialsFilePath,
		commonLabels:        nil,
		trace:               "",
		context:             "",
	}

	// Go through provided client options to configure constructor.
	for _, opt := range options {
		err := opt(conf)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	if conf.LogName == "" {
		return nil, fmt.Errorf("logname was not provided, can not create logger")
	}

	logger := &Logger{
		Logger:  nil,
		trace:   conf.trace,
		context: conf.context,
	}

	// Create logger from Google logging client.
	gcpLogger := c.Logger(conf.LogName, logging.CommonLabels(conf.commonLabels))
	logger.Logger = gcpLogger
	return logger, nil
}

// WithTrace is a logger constructor configuration option passing logger trace value to use in log entries.
func WithTrace(trace string) LoggerOption {
	return func(config *Config) error {
		config.trace = trace
		return nil
	}
}

// WithLoggerContext is a logger constructor configuration option passing logger context description value ot use in log entries.
func WithLoggerContext(context string) LoggerOption {
	return func(config *Config) error {
		config.context = context
		return nil
	}
}

// WithGeneratedTrace is a logger constructor configuration option generating logger trace value to use in log entries.
func WithGeneratedTrace() LoggerOption {
	return func(config *Config) error {
		randomInt := rand.Int()
		config.trace = fmt.Sprintf("trace/%d", randomInt)
		return nil
	}
}

// NewProwjobLogger creates logger with defaults for logging from prowjobs.
func (c *Client) NewProwjobLogger() *Logger {
	// Get default labels for prowjobs.
	prowjobLabels := GetProwjobLabels()
	l := c.Logger(ProwjobsLogName, logging.CommonLabels(prowjobLabels))
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

// getMessage format message text with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

// getEntry creates a Google logging entry.
func getEntry(severity logging.Severity, context, trace, message string, labels map[string]string) logging.Entry {
	var payloadType string
	if severity == logging.Error || severity == logging.Critical || severity == logging.Emergency {
		payloadType = ErrorReportingType
	}
	return logging.Entry{
		Timestamp: time.Now(),
		Severity:  severity,
		Labels:    labels,
		Trace:     trace,
		Payload: Payload{
			Type:    payloadType,
			Message: message,
			Context: context,
		}}
}

// getLabels converts context an array of strings in to map.
// It binds two subsequent strings in to key value pairs.
// It will return error if odd number of elements is passed in context.
func getLabels(contextLabels []string) (map[string]string, error) {
	labels := make(map[string]string)
	// Go through all
	for i := 0; i < len(contextLabels); {
		// Make sure this element isn't a dangling key. This will happen if odd number of strings is passed in contextLabels.
		if i == len(contextLabels)-1 {
			err := fmt.Errorf("an odd number of strings was passed as contextLabels, can't make key, val pairs for all elements")
			return labels, err
		}
		key, val := contextLabels[i], contextLabels[i+1]
		labels[key] = val
	}
	return labels, nil
}

// log create and send Google logging entry.
// It collects and formats all required data for Google logging entry.
func (l *Logger) log(severity logging.Severity, template string, args []interface{}, context []string) {
	var labels map[string]string
	message := getMessage(template, args)
	if context != nil {
		var err error
		labels, err = getLabels(context)
		if err != nil {
			l.Error(err.Error())
		}
	}
	entry := getEntry(severity, l.context, l.trace, message, labels)
	l.Log(entry)
}

// Error log message of Error severity.
// It will set log message payload Type to value required by gcp stackdriver to report messages as error in Error Reporting.
func (l *Logger) Error(args ...interface{}) {
	l.log(logging.Error, "", args, nil)
}

// Errorf log message of Error severity. It formats message with sprintf.
// It will set log message payload Type to value required by gcp stackdriver to report messages as error in Error Reporting.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.log(logging.Error, template, args, nil)
}

// Errorw log message of Error severity. keysAndValues is added to the log entry as key, value labels.
// It will set log message payload Type to value required by gcp stackdriver to report messages as error in Error Reporting.
func (l *Logger) Errorw(message string, keysAndValues ...string) {
	l.log(logging.Error, message, nil, keysAndValues)
}

// LogError log message of Error severity.
// It will set log message payload Type to value required by gcp stackdriver to report messages as error in Error Reporting.
func (l *Logger) LogError(message string) {
	l.log(logging.Error, message, nil, nil)
}

// Warn log message of Info severity.
func (l *Logger) Warn(args ...interface{}) {
	l.log(logging.Warning, "", args, nil)
}

// Info log message of Info severity.
func (l *Logger) Info(args ...interface{}) {
	l.log(logging.Info, "", args, nil)
}

// Infof log message of Info severity. It formats message with sprintf.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.log(logging.Info, template, args, nil)
}

// Infow log message of Info severity. keysAndValues is added to the log entry as key, value labels.
func (l *Logger) Infow(message string, keysAndValues ...string) {
	l.log(logging.Info, message, nil, keysAndValues)
}

// LogInfo log message of Info severity.
func (l *Logger) LogInfo(message string) {
	l.log(logging.Info, message, nil, nil)
}

// Debug log message of Debug severity.
func (l *Logger) Debug(args ...interface{}) {
	l.log(logging.Debug, "", args, nil)
}

// Debugf log message of Debug severity. It formats message with sprintf.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.log(logging.Debug, template, args, nil)
}

// Debugw log message of Debug severity. keysAndValues is added to the log entry as key, value labels.
func (l *Logger) Debugw(message string, keysAndValues ...string) {
	l.log(logging.Debug, message, nil, keysAndValues)
}
