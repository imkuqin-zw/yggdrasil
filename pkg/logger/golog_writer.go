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
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xcolor"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	stdDebugMsg = xcolor.Blue("DEBUG")
	stdInfoMsg  = xcolor.Green("INFO")
	stdWarnMsg  = xcolor.Yellow("WARN")
	stdErrorMsg = xcolor.Red("ERROR")
	stdFaultMsg = xcolor.Red("FAULT")
)

type WriterFile struct {
	Enable bool `json:"enable" yaml:"enable"`
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename" default:"out.log"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" yaml:"maxsize" default:"100"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage" yaml:"maxage" default:"7"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups" yaml:"maxbackups" default:"10"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `json:"localtime" yaml:"localtime"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`
}

type WriterCfg struct {
	OpenMsgFormat bool       `json:"openMsgFormat" yaml:"openMsgFormat"`
	TimeFormat    string     `json:"timeFormat" yaml:"timeFormat"`
	File          WriterFile `json:"file" yaml:"file"`
}

type writer struct {
	level      Level
	timeEncode func(t time.Time) string
	write      func(lv Level, t time.Time, msg string, ext ...[]byte)
}

func NewWriter(cfg *WriterCfg) Writer {
	w := &writer{}
	if cfg.File.Enable {
		w.write = w.newFileWrite(cfg)
	} else {
		w.write = w.newConsoleWrite(cfg)
	}
	if cfg.TimeFormat != "" {
		if cfg.File.Enable {
			w.timeEncode = func(t time.Time) string { return fmt.Sprintf(`"%s"`, t.Format(cfg.TimeFormat)) }
		} else {
			w.timeEncode = func(t time.Time) string { return t.Format(cfg.TimeFormat) }
		}
	} else {
		w.timeEncode = func(t time.Time) string { return fmt.Sprintf("%d", t.UnixMilli()) }
	}
	return w
}

func (l *writer) newFileWrite(cfg *WriterCfg) func(lv Level, t time.Time, msg string, ext ...[]byte) {
	ioWriter := &lumberjack.Logger{
		Filename:   cfg.File.Filename,
		MaxSize:    cfg.File.MaxSize,
		MaxBackups: cfg.File.MaxBackups,
		MaxAge:     cfg.File.MaxAge,
		LocalTime:  cfg.File.LocalTime,
		Compress:   cfg.File.Compress,
	}

	return func(lv Level, t time.Time, msg string, ext ...[]byte) {
		buf := Get()
		defer buf.Free()
		_ = buf.WriteByte('{')
		_, _ = fmt.Fprintf(buf, `"time":%s,`, l.timeEncode(t))
		_, _ = buf.WriteString(`"level":"`)
		_, _ = buf.WriteString(lv.String())
		_, _ = buf.WriteString(`",`)
		_, _ = buf.WriteString(`"msg":"`)
		_, _ = buf.WriteString(msg)
		_, _ = buf.WriteString(`"`)
		if len(ext) > 0 {
			_ = buf.WriteByte(',')
			_, _ = buf.Write(ext[0])
		}
		_ = buf.WriteByte('}')
		_, _ = fmt.Fprintln(ioWriter, buf.String())
	}
}

func (l *writer) newConsoleWrite(cfg *WriterCfg) func(lv Level, t time.Time, msg string, ext ...[]byte) {
	var kvsMsgFormat string
	if runtime.GOOS == "windows" {
		kvsMsgFormat = " %-8s"
	} else {
		kvsMsgFormat = "%-18s"
	}
	if cfg.OpenMsgFormat {
		kvsMsgFormat += " %-31s "
	} else {
		kvsMsgFormat += "%s "
	}
	return func(lv Level, t time.Time, msg string, ext ...[]byte) {
		buf := Get()
		defer buf.Free()
		_, _ = fmt.Fprint(buf, l.timeEncode(t), " ")
		_, _ = fmt.Fprintf(buf, kvsMsgFormat, l.getLvMsg(lv), msg)
		if len(ext) > 0 {
			_ = buf.WriteByte('{')
			_, _ = buf.Write(ext[0])
			_ = buf.WriteByte('}')
		}
		_ = buf.WriteByte('\n')
		_, _ = os.Stderr.Write(buf.Bytes())
	}
}

func (l *writer) Write(lv Level, t time.Time, msg string, ext ...[]byte) {
	l.write(lv, t, msg, ext...)
}

func (l *writer) getLvMsg(lv Level) string {
	switch lv {
	case LvDebug:
		return stdDebugMsg
	case LvInfo:
		return stdInfoMsg
	case LvWarn:
		return stdWarnMsg
	case LvError:
		return stdErrorMsg
	case LvFault:
		return stdFaultMsg
	}
	return ""
}
