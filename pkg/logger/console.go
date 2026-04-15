package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ConsoleLogger wraps zap.SugaredLogger to implement Logger.
// It outputs GCP-compatible JSON to stdout (info and below) and stderr (errors).
type ConsoleLogger struct {
	*zap.SugaredLogger
}

// Compile-time check: ConsoleLogger must implement Logger.
var _ Logger = (*ConsoleLogger)(nil)

// With creates a child logger with additional context fields.
func (l *ConsoleLogger) With(args ...interface{}) Logger {
	return &ConsoleLogger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}

// NewConsoleLogger creates a GCP-compatible console logger.
// level sets the minimum log severity (e.g. zapcore.InfoLevel, zapcore.DebugLevel).
//
// Log routing:
//   - severity >= Error → stderr
//   - severity < Error  → stdout
func NewConsoleLogger(level zapcore.Level) *ConsoleLogger {
	core := newConsoleCore(level)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &ConsoleLogger{
		SugaredLogger: zapLogger.Sugar(),
	}
}

// newConsoleCore builds a consoleCore that writes GCP-compatible structured JSON
// to stdout (info and below) and stderr (errors).
// Labels are nested under "logging.googleapis.com/labels" so Cloud Run agents
// extract them as proper Cloud Logging labels — identical structure to the API logger.
func newConsoleCore(level zapcore.Level) *consoleCore {
	return &consoleCore{
		level:  level,
		out:    zapcore.Lock(os.Stdout),
		errOut: zapcore.Lock(os.Stderr),
	}
}

// consoleCore is a custom zapcore.Core that writes GCP-compatible structured JSON
// to stdout/stderr. Labels are output as a nested "logging.googleapis.com/labels"
// object so Cloud Run log agents recognise them as Cloud Logging labels.
//
// This mirrors the structure produced by gcpCore — both loggers emit identical JSON.
type consoleCore struct {
	level  zapcore.Level
	fields []zapcore.Field
	out    zapcore.WriteSyncer
	errOut zapcore.WriteSyncer
}

func (c *consoleCore) Enabled(level zapcore.Level) bool {
	return level >= c.level
}

func (c *consoleCore) With(fields []zapcore.Field) zapcore.Core {
	return &consoleCore{
		level:  c.level,
		fields: append(append([]zapcore.Field{}, c.fields...), fields...),
		out:    c.out,
		errOut: c.errOut,
	}
}

func (c *consoleCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

// Write converts a zap log entry into a GCP-compatible JSON line.
// It mirrors the logic in gcpCore.Write so both loggers produce identical structure.
func (c *consoleCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	allFields := append(c.fields, fields...)

	enc := zapcore.NewMapObjectEncoder()
	for _, field := range allFields {
		field.AddTo(enc)
	}

	// Separate labels from regular fields.
	labels := make(map[string]interface{})
	severity, err := consoleSeverityString(entry.Level)
	if err != nil {
		return err
	}
	payload := map[string]interface{}{
		"severity":  severity,
		"timestamp": entry.Time.Format(time.RFC3339Nano),
		"message":   entry.Message,
	}

	for key, val := range enc.Fields {
		if strings.HasPrefix(key, "labels.") {
			labels[strings.TrimPrefix(key, "labels.")] = fmt.Sprint(val)
			continue
		}
		payload[key] = val
	}

	// Nest labels under "logging.googleapis.com/labels" so Cloud Run agents
	// extract them as proper Cloud Logging labels.
	if len(labels) > 0 {
		payload["logging.googleapis.com/labels"] = labels
	}

	if entry.Caller.Defined {
		payload["logging.googleapis.com/sourceLocation"] = map[string]interface{}{
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
			"function": entry.Caller.Function,
		}
	}

	if entry.Stack != "" {
		payload["stacktrace"] = entry.Stack
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if entry.Level >= zapcore.ErrorLevel {
		_, err = c.errOut.Write(data)
	} else {
		_, err = c.out.Write(data)
	}
	return err
}

func (c *consoleCore) Sync() error {
	outErr := c.out.Sync()
	errOutErr := c.errOut.Sync()
	if outErr != nil && !errors.Is(outErr, os.ErrInvalid) {
		return outErr
	}
	if errOutErr != nil && !errors.Is(errOutErr, os.ErrInvalid) {
		return errOutErr
	}
	return nil
}

// consoleSeverityString maps zap levels to Cloud Logging severity strings.
// Returns an error if the level is not supported.
func consoleSeverityString(level zapcore.Level) (string, error) {
	switch level {
	case zapcore.DebugLevel:
		return "DEBUG", nil
	case zapcore.InfoLevel:
		return "INFO", nil
	case zapcore.WarnLevel:
		return "WARNING", nil
	case zapcore.ErrorLevel:
		return "ERROR", nil
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return "CRITICAL", nil
	default:
		return "", fmt.Errorf("unsupported log level: %v", level)
	}
}
