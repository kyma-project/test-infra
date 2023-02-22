package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger return zap sugaredlogger with two output targets.
// All logs with severity Error or higher will be sent to stderr.
// All logs with severity lower than Error will be sed to stdout.
// This allows gcp logging correctly recognize log message severity.
// It implements test-infra/development/logging/LoggerInterface
// Logger has set Debug logging level.
func NewLogger() *zap.SugaredLogger {
	logger, _ := newLogger(zapcore.DebugLevel)
	return logger
}

// NewLoggerWithLevel return zap sugaredlogger with two output targets.
// All logs with severity Error or higher will be sent to stderr.
// All logs with severity lower than Error will be sed to stdout.
// This allows gcp logging correctly recognize log message severity.
// It implements test-infra/development/logging/LoggerInterface
// A zap.AtomicLevel object set logging level for logger and all downstream loggers.
// Default logging level is Info.
func NewLoggerWithLevel() (*zap.SugaredLogger, zap.AtomicLevel) {
	return newLogger(zapcore.InfoLevel)
}

// newLogger construct logger instance. Together with logger it return zap.AtomicLevel object to set logging level.
func newLogger(l zapcore.Level) (*zap.SugaredLogger, zap.AtomicLevel) {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(l)
	errorMessage := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && lvl >= atom.Level()
	})

	infoMessage := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= atom.Level()
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

	logger := zap.New(core)
	// zap.RedirectStdLog(logger)
	return logger.Sugar(), atom
}
