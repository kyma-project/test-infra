package logging

type LoggerInterface interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}
