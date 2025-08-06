package logging

import (
	"bytes"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BufferLogger wraps a zap.SugaredLogger and implements the existing
// LoggerInterface to capture logs for testing purposes.
type BufferLogger struct {
	Buffer  *bytes.Buffer
	sugared *zap.SugaredLogger
}

// NewBufferLogger creates a new logger that writes to an in-memory buffer.
func NewBufferLogger() *BufferLogger {
	var buf bytes.Buffer
	writer := zapcore.AddSync(&buf)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "" // Remove timestamp for easier assertions

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, zapcore.DebugLevel)
	logger := zap.New(core).Sugar()

	return &BufferLogger{
		Buffer:  &buf,
		sugared: logger,
	}
}

func (bl *BufferLogger) Info(args ...interface{}) {
	bl.sugared.Info(args...)
}

func (bl *BufferLogger) Warn(args ...interface{}) {
	bl.sugared.Warn(args...)
}

func (bl *BufferLogger) Error(args ...interface{}) {
	bl.sugared.Error(args...)
}

func (bl *BufferLogger) Debug(args ...interface{}) {
	bl.sugared.Debug(args...)
}

func (bl *BufferLogger) Infof(template string, args ...interface{}) {
	bl.sugared.Infof(template, args...)
}

func (bl *BufferLogger) Warnf(template string, args ...interface{}) {
	bl.sugared.Warnf(template, args...)
}

func (bl *BufferLogger) Errorf(template string, args ...interface{}) {
	bl.sugared.Errorf(template, args...)
}

func (bl *BufferLogger) Debugf(template string, args ...interface{}) {
	bl.sugared.Debugf(template, args...)
}

func (bl *BufferLogger) Infow(message string, keysAndValues ...interface{}) {
	bl.sugared.Infow(message, keysAndValues...)
}

func (bl *BufferLogger) Warnw(message string, keysAndValues ...interface{}) {
	bl.sugared.Warnw(message, keysAndValues...)
}

func (bl *BufferLogger) Errorw(message string, keysAndValues ...interface{}) {
	bl.sugared.Errorw(message, keysAndValues...)
}

func (bl *BufferLogger) Debugw(message string, keysAndValues ...interface{}) {
	bl.sugared.Debugw(message, keysAndValues...)
}

func (bl *BufferLogger) With(args ...interface{}) *zap.SugaredLogger {
	return bl.sugared.With(args...)
}

// Logs returns the captured logs as a string.
func (bl *BufferLogger) Logs() string {
	return bl.Buffer.String()
}

// Sync flushes any buffered log entries.
func (bl *BufferLogger) Sync() error {
	return bl.sugared.Sync()
}

// Ensure BufferLogger satisfies the existing LoggerInterface.
var _ LoggerInterface = (*BufferLogger)(nil)
