package logging

// LoggerInterface is a Logger interface for all implementations in kyma-project/test-infra
type LoggerInterface interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Debugw(message string, keysAndValues ...interface{})
}

// Logger is an interface to interact with Google logging.
// TODO: this should be replaced by extending LoggerInterface
type Logger interface {
	LogCritical(string)
	LogError(string)
	LogInfo(string)
}
