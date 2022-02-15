package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger return zap sugaredlogger with two output targets. All logs with severity Error or higher will be sent to stderr.
// All logs with severity lower than Error will be sed to stdout. This allows gcp logging correctly recognize log message severity.
func NewLogger() *zap.SugaredLogger {
	errorMessage := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	infoMessage := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	consoleInfo := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, errorMessage),
		zapcore.NewCore(consoleEncoder, consoleInfo, infoMessage),
	)

	return zap.New(core).Sugar()
}
