package logging

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
