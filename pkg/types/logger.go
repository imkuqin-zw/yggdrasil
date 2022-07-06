package types

type Level int8

const (
	LvDebug Level = iota - 1
	LvInfo
	LvWarn
	LvError
	LvFault
)

type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Fatal(args ...interface{})

	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	SetLevel(Level)
	GetLevel() Level
	Enable(Level) bool
}
