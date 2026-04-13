package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"go.uber.org/zap/zapcore"
)

// Environment variable names used to configure the logger.
const (
	// EnvLogDestination controls where logs are sent.
	// Values: "console", "api", "console-and-api", "auto" (default: "auto").
	EnvLogDestination = "LOG_DESTINATION"

	// EnvLogLevel controls the minimum log severity.
	// Values: "debug", "info" (default: "info").
	EnvLogLevel = "LOG_LEVEL"

	// EnvGCPProjectID is the GCP project to send logs to.
	// Required when LOG_DESTINATION is "api".
	EnvGCPProjectID = "GCP_PROJECT_ID"

	// EnvGCPLogName is the log name in Cloud Logging.
	// Optional, defaults to "application".
	EnvGCPLogName = "GCP_LOG_NAME"
)

// New creates a logger based on environment variables.
// This is the main entry point — use this in your applications:
//
//	logger, err := logger.New()
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Sync()
func New() (LoggerInterface, error) {
	level, err := parseLogLevel()
	if err != nil {
		return nil, err
	}

	destination, err := parseDestination()
	if err != nil {
		return nil, err
	}

	switch destination {
	case "console":
		return NewConsoleLogger(level).With(LogLabel("team", "neighbors-team")), nil
	case "api":
		l, err := newAPILogger(level)
		if err != nil {
			return nil, err
		}
		return l.With(LogLabel("team", "neighbors-team")), nil
	case "console-and-api":
		l, err := newCombinedLogger(level)
		if err != nil {
			return nil, err
		}
		return l.With(LogLabel("team", "neighbors-team")), nil
	default:
		return nil, fmt.Errorf("unknown log destination: %s", destination)
	}
}

// newAPILogger creates a GCP API-only logger from env vars.
func newAPILogger(level zapcore.Level) (LoggerInterface, error) {
	projectID := os.Getenv(EnvGCPProjectID)
	if projectID == "" {
		return nil, fmt.Errorf("%s is required when %s is %q", EnvGCPProjectID, EnvLogDestination, "api")
	}
	logName := os.Getenv(EnvGCPLogName)
	if logName == "" {
		logName = "application"
	}
	taskID, _ := os.Hostname()
	if taskID == "" {
		taskID = "unknown"
	}
	return NewGCPLogger(context.Background(), projectID, logName, taskID, level)
}

// newCombinedLogger creates a logger that writes to both console and GCP API
// simultaneously. Each log entry goes to stdout/stderr AND to Cloud Logging.
func newCombinedLogger(level zapcore.Level) (LoggerInterface, error) {
	projectID := os.Getenv(EnvGCPProjectID)
	if projectID == "" {
		return nil, fmt.Errorf("%s is required when %s is %q", EnvGCPProjectID, EnvLogDestination, "console-and-api")
	}
	logName := os.Getenv(EnvGCPLogName)
	if logName == "" {
		logName = "application"
	}
	taskID, _ := os.Hostname()
	if taskID == "" {
		taskID = "unknown"
	}
	return NewCombinedLogger(context.Background(), projectID, logName, taskID, level)
}

// parseDestination reads LOG_DESTINATION and resolves "auto" to a concrete value.
//
// "auto" logic:
//   - Calls metadata.OnGCE() which makes an HTTP request to 169.254.169.254
//   - If reachable → we're inside GCP → use "console" (agent collects stdout)
//   - If not reachable → we're outside GCP → use "console-and-api" (console + API simultaneously)
func parseDestination() (string, error) {
	dest := strings.ToLower(strings.TrimSpace(os.Getenv(EnvLogDestination)))

	switch dest {
	case "console":
		return "console", nil
	case "api":
		return "api", nil
	case "console-and-api":
		return "console-and-api", nil
	case "auto", "":
		if metadata.OnGCE() {
			return "console", nil
		}
		return "console-and-api", nil
	default:
		return "", fmt.Errorf(
			"invalid %s value: %q (valid: console, api, console-and-api, auto)",
			EnvLogDestination, dest,
		)
	}
}

// parseLogLevel reads LOG_LEVEL and converts it to a zapcore.Level.
func parseLogLevel() (zapcore.Level, error) {
	lvl := strings.ToLower(strings.TrimSpace(os.Getenv(EnvLogLevel)))

	switch lvl {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info", "":
		return zapcore.InfoLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf(
			"invalid %s value: %q (valid: debug, info)",
			EnvLogLevel, lvl,
		)
	}
}
