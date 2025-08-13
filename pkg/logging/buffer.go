package logging

import (
	"bytes"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BufferLogger wraps a zap.SugaredLogger and implements the existing
// LoggerInterface to capture logs for testing purposes.
type BufferLogger struct {
	*zap.SugaredLogger
	Buffer *bytes.Buffer
}

// NewBufferLogger creates a new logger that writes to an in-memory buffer.
func NewBufferLogger() *BufferLogger {
	var buf bytes.Buffer
	writer := zapcore.AddSync(&buf)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "" // Remove timestamp for easier assertions

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, zapcore.DebugLevel)
	sugaredLogger := zap.New(core).Sugar()

	return &BufferLogger{
		SugaredLogger: sugaredLogger,
		Buffer:        &buf,
	}
}

// Logs returns the captured logs as a string.
func (bl *BufferLogger) Logs() string {
	return bl.Buffer.String()
}

// Ensure BufferLogger satisfies the existing LoggerInterface.
var _ LoggerInterface = (*BufferLogger)(nil)
