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
	// Default project to store logs.
	LogsProjectID       = "sap-kyma-prow"
	CredentialsFilePath = "/etc/gcpLoggingServiceAccountKey/key"
	// Default google cloud logging log file name for log messages send by prowjobs.
	ProwjobsLogName          = "prowjobs"
	RepoOwnserServiceLogName = "repoowners"
)

// NewClient create kyma implementation of gcp logging client.
// It requires credentials file path to authenticate in GCP.
func newClient(ctx context.Context, credentialsFilePath, projectID string) (*Client, error) {
	client := &Client{}
	c, err := logging.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsFilePath))
	if err != nil {
		return nil, fmt.Errorf("got error while creating google cloud logging client, error: %w", err)
	}
	client.Client = c
	return client, nil
}

// NewGCPClient create kyma implementation of gcp logging client.
// It requires credentials file path to authenticate in GCP.
func NewGCPClient(ctx context.Context, options ...ClientOption) (*logging.Client, error) {
	conf := &Config{
		AppName:             "",
		LogName:             "",
		Component:           "",
		ProjectID:           LogsProjectID,
		credentialsFilePath: CredentialsFilePath,
	}

	for _, opt := range options {
		err := opt(conf)
		if err != nil {
			return nil, fmt.Errorf("failed applying functional option: %w", err)
		}
	}

	client, err := logging.NewClient(ctx, conf.ProjectID, option.WithCredentialsFile(conf.credentialsFilePath))
	if err != nil {
		return nil, fmt.Errorf("got error while creating google cloud logging client, error: %w", err)
	}
	return client, nil
}

func ClientWithProjectID(projectID string) ClientOption {
	return func(conf *Config) error {
		conf.ProjectID = projectID
		return nil
	}
}

func ClientWithCredentialsFilePath(credentialsFilePath string) ClientOption {
	return func(conf *Config) error {
		conf.credentialsFilePath = credentialsFilePath
		return nil
	}
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

// NewLogger creates logger with defaults for logging from prowjobs.
func NewGCPLogger(ctx context.Context, options ...LoggerOption) (*Logger, error) {
	conf := &Config{
		AppName:             "",
		LogName:             "",
		Component:           "",
		ProjectID:           LogsProjectID,
		credentialsFilePath: CredentialsFilePath,
		commonLabels:        nil,
		trace:               "",
		context:             "",
	}

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

	gcpClient, err := NewGCPClient(ctx, ClientWithProjectID(conf.ProjectID), ClientWithCredentialsFilePath(conf.credentialsFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed creating google logging client: %w", err)
	}
	gcpLogger := gcpClient.Logger(conf.LogName, logging.CommonLabels(conf.commonLabels))
	logger.Logger = gcpLogger
	return logger, nil
}

func LoggerFromConfig(config Config) LoggerOption {
	return func(conf *Config) error {
		if config.LogName == "" {
			return fmt.Errorf("Config.LogName can not be empty")
		} else {
			conf.LogName = config.LogName
		}
		if config.credentialsFilePath != "" {
			conf.credentialsFilePath = config.credentialsFilePath
		}
		if config.ProjectID != "" {
			conf.ProjectID = config.ProjectID
		}
		if config.AppName != "" {
			if conf.commonLabels == nil {
				conf.commonLabels = make(map[string]string)
			}
			conf.commonLabels["application"] = config.AppName
		}
		if config.Component != "" {
			if conf.commonLabels == nil {
				conf.commonLabels = make(map[string]string)
			}
			conf.commonLabels["component"] = config.Component
		}
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

// getMessage format with Sprint, Sprintf, or neither.
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

func getEntry(severity logging.Severity, context, trace, message string, labels map[string]string) logging.Entry {
	entry := logging.Entry{
		Timestamp: time.Now(),
		Severity:  severity,
	}
	payload := Payload{
		Message: message,
	}
	if severity == logging.Error || severity == logging.Critical || severity == logging.Emergency {
		payload.Type = ErrorReportingType
	}
	if context != "" {
		payload.Context = context
	}
	entry.Payload = payload
	if trace != "" {
		entry.Trace = trace
	}
	if labels != nil {
		entry.Labels = labels
	}
	return entry
}

// TODO: write implementation
func getLabels(context []interface{}) map[string]string {
	return nil
}

func (l *Logger) log(severity logging.Severity, template string, args, context []interface{}) {
	message := getMessage(template, args)
	labels := getLabels(context)
	entry := getEntry(severity, l.context, l.trace, message, labels)
	l.Log(entry)
}

// LogError log message of Error severity with defaults. Error message is send to standard logger as well
// It will set log message payload Type to value required by gcp stackdriver to report messages as error in Error Reporting.
func (l *Logger) Error(args ...interface{}) {
	l.log(logging.Error, "", args, nil)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.log(logging.Error, template, args, nil)
}

func (l *Logger) LogError(message string) {
	l.log(logging.Error, message, nil, nil)
}

// LogInfo log message of Info severity wit defaults. Info message is send to standard logger as well.
func (l *Logger) Info(args ...interface{}) {
	l.log(logging.Info, "", args, nil)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.log(logging.Info, template, args, nil)
}

func (l *Logger) LogInfo(message string) {
	l.log(logging.Info, message, nil, nil)
}
