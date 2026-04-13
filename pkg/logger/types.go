package logger

import (
	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
)

// LoggerInterface is the primary logger interface.
// Accept this interface in all function and method signatures that require logging.
// It composes structured logging, child logger creation, and log flushing.
type LoggerInterface interface {
	StructuredLoggerInterface
	WithLoggerInterface
	SyncLoggerInterface
}

// StructuredLoggerInterface defines structured logging methods.
// All log messages use key-value pairs for structured data.
// The "w" suffix stands for "with" — each method accepts a message
// followed by alternating key-value pairs:
//
//	logger.Infow("user logged in", "user_id", "123", "ip", "10.0.0.1")
type StructuredLoggerInterface interface {
	Infow(message string, keysAndValues ...interface{})
	Warnw(message string, keysAndValues ...interface{})
	Errorw(message string, keysAndValues ...interface{})
	Debugw(message string, keysAndValues ...interface{})
}

// WithLoggerInterface defines a method to create a child logger
// with additional context fields. Every log entry from the child logger
// will include these fields automatically.
//
// Example:
//
//	child := logger.With("request_id", "abc-123", "component", "auth")
//	child.Infow("processing request")  // includes request_id and component
//
// Returns LoggerInterface (not a concrete type) to maintain abstraction.
type WithLoggerInterface interface {
	With(args ...interface{}) LoggerInterface
}

// SyncLoggerInterface defines a method to flush buffered log entries.
// Call Sync before application exit to ensure all logs are delivered.
type SyncLoggerInterface interface {
	Sync() error
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
