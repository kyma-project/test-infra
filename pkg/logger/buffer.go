package logger

import (
	"bytes"

	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BufferLogger writes logs to an in-memory buffer for testing.
// It uses the same GCP-compatible JSON format as ConsoleLogger,
// so tests validate the exact output format that production uses.
type BufferLogger struct {
	*zap.SugaredLogger
	// Buffer holds all log output. Read it with Logs() method.
	Buffer *bytes.Buffer
}

// Compile-time check: BufferLogger must implement LoggerInterface.
var _ LoggerInterface = (*BufferLogger)(nil)

// With creates a child logger with additional context fields.
// Same override as ConsoleLogger — wraps the returned SugaredLogger
// back into BufferLogger so it still writes to the same buffer.
func (l *BufferLogger) With(args ...interface{}) LoggerInterface {
	return &BufferLogger{
		SugaredLogger: l.SugaredLogger.With(args...),
		Buffer:        l.Buffer,
	}
}

// Logs returns all captured log output as a string.
func (l *BufferLogger) Logs() string {
	return l.Buffer.String()
}

// NewBufferLogger creates a logger that captures output in memory.
// Uses the same JSON encoder config as ConsoleLogger (zapdriver)
// so the format matches production output.
func NewBufferLogger() *BufferLogger {
	var buf bytes.Buffer

	// Same encoder as production — tests validate real output format.
	encoderConfig := zapdriver.NewProductionEncoderConfig()
	// Remove timestamp — makes test assertions deterministic.
	// Without this, every log line has a different timestamp.
	encoderConfig.TimeKey = ""

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// AddSync wraps bytes.Buffer into a WriteSyncer.
	// bytes.Buffer implements io.Writer but not Sync(),
	// so AddSync adds a no-op Sync() method.
	writer := zapcore.AddSync(&buf)

	// DebugLevel — capture everything in tests, don't filter.
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)

	zapLogger := zap.New(core)

	return &BufferLogger{
		SugaredLogger: zapLogger.Sugar(),
		Buffer:        &buf,
	}
}
