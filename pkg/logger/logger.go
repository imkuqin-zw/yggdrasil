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
	"sync"
	"time"
)

var _fieldsPool = newFieldsPool()

type fieldsPool struct {
	pool sync.Pool
}

func newFieldsPool() *fieldsPool {
	return &fieldsPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]Field, 0, 10) // 10为数组的初始容量
			},
		},
	}
}

func (ap *fieldsPool) Get() []Field {
	return ap.pool.Get().([]Field)
}

func (ap *fieldsPool) Put(arr []Field) {
	arr = arr[:0] // 重置切片长度
	ap.pool.Put(arr)
}

type Logger struct {
	lv     *Level
	writer *Writer
	fields []Field
}

func NewLogger(lv Level, writer Writer) *Logger {
	return &Logger{lv: &lv, writer: &writer}
}

func (l *Logger) Debug(args ...interface{}) {
	l.out(LvDebug, "", args)
}

func (l *Logger) Info(args ...interface{}) {
	l.out(LvInfo, "", args)
}

func (l *Logger) Warn(args ...interface{}) {
	l.out(LvWarn, "", args)
}

func (l *Logger) Error(args ...interface{}) {
	l.out(LvError, "", args)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.out(LvFault, "", args)
	os.Exit(1)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.out(LvDebug, format, args)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.out(LvInfo, format, args)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.out(LvWarn, format, args)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.out(LvError, format, args)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.out(LvFault, format, args)
	os.Exit(1)
}

func (l *Logger) DebugField(msg string, fields ...Field) {
	l.out(LvDebug, msg, nil, fields...)
}

func (l *Logger) InfoField(msg string, fields ...Field) {
	l.out(LvInfo, msg, nil, fields...)
}

func (l *Logger) WarnField(msg string, fields ...Field) {
	l.out(LvWarn, msg, nil, fields...)
}

func (l *Logger) ErrorField(msg string, fields ...Field) {
	l.out(LvError, msg, nil, fields...)
	l.printStack()
}

func (l *Logger) FatalField(msg string, fields ...Field) {
	l.out(LvFault, msg, nil, fields...)
	l.printStack()
	os.Exit(1)
}

func (l *Logger) SetLevel(level Level) {
	*l.lv = level
}

func (l *Logger) GetLevel() Level {
	return *l.lv
}

func (l *Logger) Enable(level Level) bool {
	return l.lv.Enable(level)
}

func (l *Logger) SetWriter(w Writer) {
	*l.writer = w
}

func (l *Logger) Clone() *Logger {
	fields := make([]Field, len(l.fields))
	copy(fields, l.fields)
	var lv Level
	_ = lv.UnmarshalText([]byte(l.lv.String()))
	return &Logger{lv: &lv, writer: l.writer, fields: fields}
}

func (l *Logger) WithFields(fields ...Field) *Logger {
	fields = ignoreSkip(fields)
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)
	newHelper := *l
	newHelper.fields = newFields
	return &newHelper
}

func (l *Logger) mergeFields(fields ...Field) []Field {
	fields = ignoreSkip(fields)
	final := _fieldsPool.Get()
	return append(append(final, l.fields...), fields...)
}

func (l *Logger) out(lv Level, format string, args []interface{}, fields ...Field) {
	l.write(lv, time.Now(), format, args, fields...)
}

func (l *Logger) write(lv Level, t time.Time, format string, args []interface{}, fields ...Field) {
	if !l.Enable(lv) {
		return
	}
	msgBuf := Get()
	defer msgBuf.Free()
	if len(format) > 0 {
		_, _ = fmt.Fprintf(msgBuf, format, args...)
	} else {
		_, _ = fmt.Fprint(msgBuf, args...)
	}

	if len(l.fields) > 0 {
		fields = l.mergeFields(fields...)
		defer _fieldsPool.Put(fields)
	}
	if len(fields) == 0 {
		(*l.writer).Write(lv, t, msgBuf.String())
		return
	}
	fieldsBuf, _ := enc.Encode(fields)
	defer fieldsBuf.Free()
	if fieldsBuf.Len() > 0 {
		(*l.writer).Write(lv, t, msgBuf.String(), fieldsBuf.Bytes())
	} else {
		(*l.writer).Write(lv, t, msgBuf.String())
	}

}

func (l *Logger) printStack(fields ...Field) {
	if printStack {
		for _, item := range fields {
			if item.Type == ErrorType {
				if formatter, ok := item.Interface.(fmt.Formatter); ok {
					_, _ = fmt.Fprintf(os.Stderr, "%+v\n", formatter)
				}
				return
			}
		}
	}
}
