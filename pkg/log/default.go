package log

import (
	"log"
	"os"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

var lg types.Logger

func init() {
	lg = &StdLogger{level: types.LvDebug, lg: log.Default()}
}

func SetLogger(logger types.Logger) {
	lg = logger
}

// Debug is debug level
func Debug(args ...interface{}) {
	lg.Debug(args...)
}

// Info is info level
func Info(args ...interface{}) {
	lg.Info(args...)
}

// Warn is warning level
func Warn(args ...interface{}) {
	lg.Warn(args...)
}

// Error is error level
func Error(args ...interface{}) {
	lg.Error(args...)
}

// Error is fault level
func Fatal(args ...interface{}) {
	lg.Fatal(args...)
	os.Exit(1)
}

// Debugf is format debug level
func Debugf(fmt string, args ...interface{}) {
	lg.Debugf(fmt, args...)
}

// Infof is format info level
func Infof(fmt string, args ...interface{}) {
	lg.Infof(fmt, args...)
}

// Warnf is format warning level
func Warnf(fmt string, args ...interface{}) {
	lg.Warnf(fmt, args...)
}

// Errorf is format error level
func Errorf(fmt string, args ...interface{}) {
	lg.Errorf(fmt, args...)
}

// Fatalf is format fatal level
func Fatalf(fmt string, args ...interface{}) {
	lg.Fatalf(fmt, args...)
	os.Exit(1)
}

func SetLevel(level types.Level) {
	lg.SetLevel(level)
}

func GetLevel() types.Level {
	return lg.GetLevel()
}

func Enable(level types.Level) bool {
	return lg.Enable(level)
}

func GetRaw() interface{} {
	return lg
}
