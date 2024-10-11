package logging

import (
	"go.uber.org/zap"
)

// LoggerInterface is a Logger interface collecting other main Logger interfaces.
// It is used for compatibility with existing implementations in kyma-project/test-infra
type LoggerInterface interface {
	SimpleLoggerInterface
	FormatedLoggerInterface
	StructuredLoggerInterface
}

// SimpleLoggerInterface is a Logger interface with simple logging methods.
type SimpleLoggerInterface interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
}

// FormatedLoggerInterface is a Logger interface with logging methods supporting formated logging messages.
type FormatedLoggerInterface interface {
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Debugf(template string, args ...interface{})
}

// StructuredLoggerInterface is a Logger interface with logging methods supporting structured logging.
type StructuredLoggerInterface interface {
	Infow(message string, keysAndValues ...interface{})
	Warnw(message string, keysAndValues ...interface{})
	Errorw(message string, keysAndValues ...interface{})
	Debugw(message string, keysAndValues ...interface{})
}

// WithLoggerInterface is a Logger interface with support for adding context fields to the logger.
type WithLoggerInterface interface {
	With(args ...interface{}) *zap.SugaredLogger
}

// Logger is an interface to interact with Google logging.
// TODO: this should be replaced by extending LoggerInterface
type Logger interface {
	LogCritical(string)
	LogError(string)
	LogInfo(string)
}
