package log

import (
	"fmt"
	"log"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xcolor"
)

type StdLogger struct {
	level types.Level
	lg    *log.Logger
}

func (l *StdLogger) Info(args ...interface{}) {
	if l.Enable(types.LvInfo) {
		l.lg.Print(fmt.Sprintf("%s   %s", xcolor.Green("info"), fmt.Sprintln(args...)))
	}
}

func (l *StdLogger) Warn(args ...interface{}) {
	if l.Enable(types.LvWarn) {
		l.lg.Print(fmt.Sprintf("%s   %s", xcolor.Yellow("warn"), fmt.Sprint(args...)))
	}
}

func (l *StdLogger) Error(args ...interface{}) {
	if l.Enable(types.LvError) {
		l.lg.Print(fmt.Sprintf("%s   %s", xcolor.Red("error"), fmt.Sprint(args...)))
	}
}

func (l *StdLogger) Debug(args ...interface{}) {
	if l.Enable(types.LvDebug) {
		l.lg.Print(fmt.Sprintf("%s   %s", xcolor.Blue("debug"), fmt.Sprint(args...)))
	}
}

func (l *StdLogger) Fatal(args ...interface{}) {
	if l.Enable(types.LvFault) {
		l.lg.Println(fmt.Sprintf("%s   %s", xcolor.Red("fault"), fmt.Sprintln(args...)))
	}
}

func (l *StdLogger) Infof(format string, args ...interface{}) {
	if l.Enable(types.LvInfo) {
		l.lg.Println(fmt.Sprintf("%s   %s", xcolor.Green("info"), fmt.Sprintf(format, args...)))
	}
}

func (l *StdLogger) Warnf(format string, args ...interface{}) {
	if l.Enable(types.LvWarn) {
		l.lg.Println(fmt.Sprintf("%s   %s", xcolor.Yellow("warn"), fmt.Sprintf(format, args...)))
	}
}

func (l *StdLogger) Errorf(format string, args ...interface{}) {
	if l.Enable(types.LvError) {
		l.lg.Println(fmt.Sprintf("%s   %s", xcolor.Red("error"), fmt.Sprintf(format, args...)))
	}
}

func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.Enable(types.LvDebug) {
		l.lg.Println(fmt.Sprintf("%s   %s", xcolor.Blue("debug"), fmt.Sprintf(format, args...)))
	}
}

func (l *StdLogger) Fatalf(format string, args ...interface{}) {
	if l.Enable(types.LvFault) {
		l.lg.Println(fmt.Sprintf("%s   %s", xcolor.Red("fault"), fmt.Sprintf(format, args...)))
	}
}

func (l *StdLogger) SetLevel(level types.Level) {
	l.level = level
}

func (l *StdLogger) GetLevel() types.Level {
	return l.level
}

func (l *StdLogger) Enable(level types.Level) bool {
	return l.level <= level
}

func (l *StdLogger) GetRaw() interface{} {
	return l
}
