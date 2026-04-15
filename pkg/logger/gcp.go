package logger

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

// GCPLogger sends logs to Google Cloud Logging via API.
// Use this for workloads running outside GCP that need to send
// logs to the centralized logging system.
type GCPLogger struct {
	*zap.SugaredLogger
	client *logging.Client
}

// Compile-time check: GCPLogger must implement Logger.
var _ Logger = (*GCPLogger)(nil)

// With creates a child logger with additional context fields.
func (l *GCPLogger) With(args ...interface{}) Logger {
	return &GCPLogger{
		SugaredLogger: l.SugaredLogger.With(args...),
		client:        l.client,
	}
}

// Close flushes pending logs and closes the underlying GCP client.
// Call this before application exit (in addition to Sync).
func (l *GCPLogger) Close() error {
	if err := l.Sync(); err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}
	if l.client != nil {
		return l.client.Close()
	}
	return nil
}

// NewGCPLogger creates a logger that sends logs to Google Cloud Logging API.
//
// Parameters:
//   - ctx: context for creating the GCP client
//   - projectID: GCP project to send logs to (e.g. "sap-kyma-neighbors-dev")
//   - logName: log name in Cloud Logging (e.g. "application"). Under this log name, logs from every application will be
//     grouped in Cloud Logging UI.
//   - namespace: logical grouping for the workload (e.g. "neighbors-team")
//   - level: minimum log severity
func NewGCPLogger(ctx context.Context, projectID, logName, taskID string, level zapcore.Level) (*GCPLogger, error) {
	client, gcpCore, err := newGCPCore(ctx, projectID, logName, taskID, level)
	if err != nil {
		return nil, err
	}

	zapLogger := zap.New(gcpCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &GCPLogger{
		SugaredLogger: zapLogger.Sugar(),
		client:        client,
	}, nil
}

// NewCombinedLogger creates a logger that writes to both console (stdout/stderr)
// and GCP Cloud Logging API simultaneously. Used for workloads outside GCP
// that need both local visibility and centralized logging.
func NewCombinedLogger(ctx context.Context, projectID, logName, taskID string, level zapcore.Level) (*GCPLogger, error) {
	client, gcpAPICore, err := newGCPCore(ctx, projectID, logName, taskID, level)
	if err != nil {
		return nil, err
	}

	consoleCore := newConsoleCore(level)

	// Tee combines both cores — each log entry goes to both destinations.
	combined := zapcore.NewTee(consoleCore, gcpAPICore)
	zapLogger := zap.New(combined, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &GCPLogger{
		SugaredLogger: zapLogger.Sugar(),
		client:        client,
	}, nil
}

// newGCPCore creates the Cloud Logging client and a gcpCore.
// Shared by NewGCPLogger and NewCombinedLogger.
func newGCPCore(ctx context.Context, projectID, logName, taskID string, level zapcore.Level) (*logging.Client, *gcpCore, error) {
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCP logging client: %w", err)
	}

	gcpLogger := client.Logger(logName, logging.CommonResource(&monitoredres.MonitoredResource{
		Type: "generic_task",
		Labels: map[string]string{
			"project_id": projectID,
			"namespace":  "neighbors-team",
			"job":        logName,
			"task_id":    taskID,
			"location":   "global",
		},
	}))

	core := &gcpCore{
		level:     zapcore.Level(level),
		gcpLogger: gcpLogger,
	}

	return client, core, nil
}

// gcpCore is a custom zapcore.Core that sends log entries to Cloud Logging API
// instead of writing to stdout/stderr.
//
// How it works:
//  1. zap calls Write() with a log entry and fields
//  2. We collect all fields into a map
//  3. Extract labels (fields with "labels." prefix from zapdriver)
//  4. Extract trace ID if present
//  5. Map zap severity to GCP severity
//  6. Build a logging.Entry and send it via the GCP client
type gcpCore struct {
	level     zapcore.Level
	gcpLogger *logging.Logger
	fields    []zapcore.Field // fields accumulated via With()
}

// Enabled checks if the given log level should be logged.
func (c *gcpCore) Enabled(level zapcore.Level) bool {
	return level >= c.level
}

// With returns a new core with additional fields.
// Called when user does logger.With("key", "val") — the fields
// are stored and included in every subsequent log entry.
func (c *gcpCore) With(fields []zapcore.Field) zapcore.Core {
	return &gcpCore{
		level:     c.level,
		gcpLogger: c.gcpLogger,
		fields:    append(append([]zapcore.Field{}, c.fields...), fields...),
	}
}

// Check determines whether the entry should be logged.
// Standard boilerplate — just adds this core to the checked entry.
func (c *gcpCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

// Write takes a log entry from zap
// and converts it into a Cloud Logging API call.
func (c *gcpCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Merge stored fields (from With()) with per-entry fields.
	allFields := append(c.fields, fields...)

	// Use zap's MapObjectEncoder to convert fields into a map.
	// This handles all zap field types (String, Int, Bool, etc.)
	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range allFields {
		field.AddTo(encoder)
	}

	// Separate labels from regular fields.
	// zapdriver labels appear as "labels.key" in the encoder output.
	labels := make(map[string]string)
	payload := map[string]interface{}{
		"message": entry.Message,
	}

	for key, val := range encoder.Fields {
		if strings.HasPrefix(key, "labels.") {
			// "labels.app" → label key "app"
			labelKey := strings.TrimPrefix(key, "labels.")
			labels[labelKey] = fmt.Sprint(val)
			continue
		}
		// Regular field — goes into the payload.
		payload[key] = val
	}

	// Add source location to payload if available.
	if entry.Caller.Defined {
		payload["caller"] = entry.Caller.String()
	}

	// Add stacktrace if present (errors).
	if entry.Stack != "" {
		payload["stacktrace"] = entry.Stack
	}

	// Map zap severity to Cloud Logging severity.
	severity := mapSeverity(entry.Level)

	// Extract trace ID if present.
	trace := ""
	if t, ok := payload["logging.googleapis.com/trace"].(string); ok {
		trace = t
		delete(payload, "logging.googleapis.com/trace")
	}

	// Build and send the Cloud Logging entry.
	logEntry := logging.Entry{
		Timestamp: entry.Time,
		Severity:  severity,
		Payload:   payload,
		Labels:    labels,
		Trace:     trace,
	}

	c.gcpLogger.Log(logEntry)

	// Auto-flush on errors to ensure they're delivered immediately.
	if entry.Level >= zapcore.ErrorLevel {
		_ = c.Sync()
	}

	return nil
}

// Sync flushes all buffered log entries to Cloud Logging.
func (c *gcpCore) Sync() error {
	return c.gcpLogger.Flush()
}

// mapSeverity converts a zap log level to a Cloud Logging severity.
func mapSeverity(level zapcore.Level) logging.Severity {
	switch level {
	case zapcore.DebugLevel:
		return logging.Debug
	case zapcore.InfoLevel:
		return logging.Info
	case zapcore.WarnLevel:
		return logging.Warning
	case zapcore.ErrorLevel:
		return logging.Error
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return logging.Critical
	default:
		return logging.Default
	}
}
