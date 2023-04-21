// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger/buffer"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xcolor"
)

var (
	stdDebugMsg = xcolor.Blue("DEBUG")
	stdInfoMsg  = xcolor.Green("INFO")
	stdWarnMsg  = xcolor.Yellow("WARN")
	stdErrorMsg = xcolor.Red("ERROR")
	stdFaultMsg = xcolor.Red("FAULT")
)

type StdLogger struct {
	level        Level
	lg           *log.Logger
	kvsMsgFormat string
}

func NewStdLogger(lv Level, lg *log.Logger) *StdLogger {
	return &StdLogger{level: lv, lg: lg, kvsMsgFormat: lvFormat + "%s "}
}

func (l *StdLogger) OpenMsgFormat() {
	l.kvsMsgFormat = lvFormat + "%-31s "
}

func (l *StdLogger) Debug(args ...interface{}) {
	if l.Enable(LvDebug) {
		l.lg.Print(l.argsFormat(stdDebugMsg, args))
	}
}

func (l *StdLogger) Info(args ...interface{}) {
	if l.Enable(LvInfo) {
		l.lg.Print(l.argsFormat(stdInfoMsg, args))
	}
}

func (l *StdLogger) Warn(args ...interface{}) {
	if l.Enable(LvWarn) {
		l.lg.Print(l.argsFormat(stdWarnMsg, args))
	}
}

func (l *StdLogger) Error(args ...interface{}) {
	if l.Enable(LvError) {
		l.lg.Print(l.argsFormat(stdErrorMsg, args))
	}
}

func (l *StdLogger) Fatal(args ...interface{}) {
	if l.Enable(LvFault) {
		l.lg.Print(l.argsFormat(stdFaultMsg, args))
	}
}

func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.Enable(LvDebug) {
		l.lg.Print(l.tplFormat(stdDebugMsg, format, args))
	}
}

func (l *StdLogger) Infof(format string, args ...interface{}) {
	if l.Enable(LvInfo) {
		l.lg.Print(l.tplFormat(stdInfoMsg, format, args))
	}
}

func (l *StdLogger) Warnf(format string, args ...interface{}) {
	if l.Enable(LvWarn) {
		l.lg.Print(l.tplFormat(stdWarnMsg, format, args))
	}
}

func (l *StdLogger) Errorf(format string, args ...interface{}) {
	if l.Enable(LvError) {
		l.lg.Print(l.tplFormat(stdErrorMsg, format, args))
	}
}

func (l *StdLogger) Fatalf(format string, args ...interface{}) {
	if l.Enable(LvFault) {
		l.lg.Print(l.tplFormat(stdFaultMsg, format, args))
	}
}

func (l *StdLogger) Debugw(msg string, kvs ...interface{}) {
	if l.Enable(LvDebug) {
		if len(kvs) == 0 {
			l.Infof(msg)
			return
		}
		l.lg.Println(l.kvsFormat(stdDebugMsg, msg, kvs))
	}
}

func (l *StdLogger) Infow(msg string, kvs ...interface{}) {
	if l.Enable(LvInfo) {
		if len(kvs) == 0 {
			l.Infof(msg)
			return
		}
		l.lg.Println(l.kvsFormat(stdInfoMsg, msg, kvs))
	}
}

func (l *StdLogger) Warnw(msg string, kvs ...interface{}) {
	if l.Enable(LvWarn) {
		if len(kvs) == 0 {
			l.Infof(msg)
			return
		}
		l.lg.Println(l.kvsFormat(stdWarnMsg, msg, kvs))
	}
}

func (l *StdLogger) Errorw(msg string, kvs ...interface{}) {
	if l.Enable(LvError) {
		if len(kvs) == 0 {
			l.Infof(msg)
			return
		}
		l.lg.Println(l.kvsFormat(stdErrorMsg, msg, kvs))
	}
}

func (l *StdLogger) Fatalw(msg string, kvs ...interface{}) {
	if l.Enable(LvFault) {
		if len(kvs) == 0 {
			l.Infof(msg)
			return
		}
		l.lg.Println(l.kvsFormat(stdFaultMsg, msg, kvs))
	}
}

func (l *StdLogger) SetLevel(level Level) {
	l.level = level
}

func (l *StdLogger) GetLevel() Level {
	return l.level
}

func (l *StdLogger) Enable(level Level) bool {
	return l.level <= level
}

func (l *StdLogger) Clone() Logger {
	lg := *l
	return &lg
}

func (l *StdLogger) GetRaw() interface{} {
	return l
}

func paris(buf *buffer.Buffer, kvs []interface{}) string {
	if (len(kvs) & 1) == 1 {
		kvs = append(kvs, "KEYVALS UNPAIRED")
	}
	_ = buf.WriteByte('{')
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	re := defaultReflectedEncoder(buf)
	for i := 0; i < len(kvs); i += 2 {
		_, _ = buf.WriteString(fmt.Sprintf(`"%v":`, kvs[i]))
		_ = re.Encode(kvs[i+1])
		buf.TrimNewline()
	}
	_ = buf.WriteByte('}')
	return buf.String()
}

func (l *StdLogger) kvsFormat(lv string, msg string, kvs []interface{}) string {
	buf := Get()
	defer buf.Free()
	_, _ = fmt.Fprintf(buf, l.kvsMsgFormat, lv, msg)
	paris(buf, kvs)
	return buf.String()
}

func (l *StdLogger) argsFormat(lv string, args []interface{}) string {
	buf := Get()
	defer buf.Free()
	_, _ = fmt.Fprintf(buf, lvFormat, lv)
	for _, arg := range args {
		_, _ = fmt.Fprintf(buf, "%v", arg)
		_ = buf.WriteByte(' ')
	}

	return buf.String()
}

func (l *StdLogger) tplFormat(lv string, format string, args []interface{}) string {
	buf := Get()
	defer buf.Free()
	_, _ = fmt.Fprintf(buf, lvFormat, lv)
	_, _ = fmt.Fprintf(buf, format, args...)
	return buf.String()
}
