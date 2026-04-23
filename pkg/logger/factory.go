package logger

import (
	"context"
	"fmt"

	"github.com/blendle/zapdriver"
	logging "github.tools.sap/kyma/neighbors-contracts/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Environment variable names used to configure the logger.
// The caller is responsible for reading these and passing values via Config.
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

// Config holds all configuration needed to create a logger.
// The caller is responsible for populating this — typically by reading
// environment variables or flags in main().
type Config struct {
	// Level is the minimum log severity. Defaults to Info.
	Level zapcore.Level

	// Destination controls where logs are sent.
	// Valid values: "console", "api", "console-and-api". Defaults to "console".
	Destination string

	// ProjectID is the GCP project to send logs to.
	// Required when Destination is "api" or "console-and-api".
	ProjectID string

	// LogName is the log name in Cloud Logging.
	// Optional, defaults to "application".
	LogName string

	// TaskID is a unique identifier for this workload instance (e.g. hostname).
	TaskID string
}

// New creates a logger based on the provided Config.
//
// Example usage in main():
//
//	cfg := logger.Config{
//	    Level:       zapcore.InfoLevel,
//	    Destination: os.Getenv(logger.EnvLogDestination),
//	    ProjectID:   os.Getenv(logger.EnvGCPProjectID),
//	    LogName:     os.Getenv(logger.EnvGCPLogName),
//	    TaskID:      hostname,
//	}
//	l, err := logger.New(ctx, cfg)
//	if err != nil {
//	    panic(err)
//	}
//	defer l.Sync()
func New(ctx context.Context, cfg Config) (logging.LoggerInterface, error) {
	logName := cfg.LogName
	if logName == "" {
		logName = "application"
	}

	switch cfg.Destination {
	case "console", "":
		return newConsoleLogger(cfg.Level), nil
	case "api":
		if cfg.ProjectID == "" {
			return nil, fmt.Errorf("ProjectID is required when Destination is %q", cfg.Destination)
		}
		return newGCPLogger(ctx, cfg.ProjectID, logName, cfg.TaskID, cfg.Level)
	case "console-and-api":
		if cfg.ProjectID == "" {
			return nil, fmt.Errorf("ProjectID is required when Destination is %q", cfg.Destination)
		}
		return newCombinedLogger(ctx, cfg.ProjectID, logName, cfg.TaskID, cfg.Level)
	default:
		return nil, fmt.Errorf("invalid Destination %q (valid: console, api, console-and-api)", cfg.Destination)
	}
}

// LogLabel creates a GCP Cloud Logging label field.
//
// Labels are special — in GCP they land in a separate "labels" field,
// not in the JSON payload. This makes them indexed and filterable
// in Cloud Logging console.
//
// Use LogLabel for metadata like app name, version, environment.
// Use regular key-value pairs for dynamic data like request_id, user_id.
//
// Example:
//
//	appLogger := logger.With(logger.LogLabel("app", "image-builder"), logger.LogLabel("version", "1.0.0"))
func LogLabel(key, value string) zap.Field {
	return zapdriver.Label(key, value)
}
