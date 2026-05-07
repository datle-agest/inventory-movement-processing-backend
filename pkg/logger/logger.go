package logger

type ServiceLogger interface {
	InitFlags()
	Activate() error
	Stop() error

	GetLogger(prefix string) Logger
	GetLevel() string
	SetLevel(string) error
}

type Fields map[string]interface{}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	With(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
}
