package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

// Environment variable names used to configure the logger.
const (
	// EnvLogDestination controls where logs are sent.
	// Values: "console" (default), "api", "console-and-api".
	EnvLogDestination = "LOG_DESTINATION"

	// EnvLogLevel controls the minimum log severity.
	// Values: "debug", "info", "warn", "error", "dpanic", "panic", "fatal" (default: "info").
	EnvLogLevel = "LOG_LEVEL"

	// EnvGCPProjectID is the GCP project to send logs to.
	// Required when LOG_DESTINATION is "api" or "console-and-api".
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
func New() (Logger, error) {
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
func newAPILogger(level zapcore.Level) (Logger, error) {
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
func newCombinedLogger(level zapcore.Level) (Logger, error) {
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

// parseDestination reads LOG_DESTINATION.
// Defaults to "console" when not set.
func parseDestination() (string, error) {
	dest := strings.ToLower(strings.TrimSpace(os.Getenv(EnvLogDestination)))

	switch dest {
	case "console", "":
		return "console", nil
	case "api":
		return "api", nil
	case "console-and-api":
		return "console-and-api", nil
	default:
		return "", fmt.Errorf(
			"invalid %s value: %q (valid: console, api, console-and-api)",
			EnvLogDestination, dest,
		)
	}
}

// parseLogLevel reads LOG_LEVEL and converts it to a zapcore.Level.
// Defaults to info when LOG_LEVEL is not set.
func parseLogLevel() (zapcore.Level, error) {
	lvl := strings.ToLower(strings.TrimSpace(os.Getenv(EnvLogLevel)))

	switch lvl {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info", "":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "dpanic":
		return zapcore.DPanicLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf(
			"invalid %s value: %q (valid: debug, info, warn, error, dpanic, panic, fatal)",
			EnvLogLevel, lvl,
		)
	}
}
