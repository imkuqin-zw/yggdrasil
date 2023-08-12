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
	"errors"
)

var (
	global     *Logger
	enc        Encoder
	printStack bool
)

func init() {
	enc = &jsonEncoder{
		EncodeTime:     RFC3339TimeEncoder,
		EncodeDuration: MillisDurationEncoder,
		buf:            Get(),
	}
	printStack = true
	global = &Logger{lv: LvDebug, writer: NewWriter(&WriterCfg{OpenMsgFormat: true})}
}

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *Level")

func SetLogger(lg *Logger) {
	global = lg
}

func SetDurationEncoder(de DurationEncoder) {
	enc.SetDurationEncoder(de)
}

func SetTimeEncoder(te TimeEncoder) {
	enc.SetTimeEncoder(te)
}

func SetDurationEncoderByName(name string) error {
	var de DurationEncoder
	switch name {
	case "seconds":
		de = SecondsDurationEncoder
	case "nanos":
		de = NanosDurationEncoder
	case "millis":
		de = MillisDurationEncoder
	case "string":
		de = StringDurationEncoder
	default:
		return errors.New("unknown time encoder")
	}
	enc.SetDurationEncoder(de)
	return nil
}

func SetTimeEncoderByName(name string) error {
	var te TimeEncoder
	switch name {
	case "RFC3339":
		te = RFC3339TimeEncoder
	case "RFC3339Nano":
		te = RFC3339NanoTimeEncoder
	case "ISO8601":
		te = ISO8601TimeEncoder
	case "epoch":
		te = EpochTimeEncoder
	case "epochNanos":
		te = EpochNanosTimeEncoder
	case "epochMillis":
		te = EpochMillisTimeEncoder
	default:
		return errors.New("unknown time encoder")
	}
	enc.SetTimeEncoder(te)
	return nil
}

func SetStackPrintState(b bool) {
	printStack = b
}

func Debug(args ...interface{}) {
	global.Debug(args...)
}

func Info(args ...interface{}) {
	global.Info(args...)
}

func Warn(args ...interface{}) {
	global.Warn(args...)
}

func Error(args ...interface{}) {
	global.Error(args...)
}

func Fatal(args ...interface{}) {
	global.Fatal(args...)
}

func Debugf(fmt string, args ...interface{}) {
	global.Debugf(fmt, args...)
}

func Infof(fmt string, args ...interface{}) {
	global.Infof(fmt, args...)
}

func Warnf(fmt string, args ...interface{}) {
	global.Warnf(fmt, args...)
}

func Errorf(fmt string, args ...interface{}) {
	global.Errorf(fmt, args...)
}

func Fatalf(fmt string, args ...interface{}) {
	global.Fatalf(fmt, args...)
}

func DebugField(msg string, fields ...Field) {
	global.DebugField(msg, fields...)
}

func InfoField(msg string, fields ...Field) {
	global.InfoField(msg, fields...)
}

func WarnField(msg string, fields ...Field) {
	global.WarnField(msg, fields...)
}

func ErrorField(msg string, fields ...Field) {
	global.ErrorField(msg, fields...)
}

func FatalField(msg string, fields ...Field) {
	global.FatalField(msg, fields...)
}

func WithFields(fields ...Field) *Logger {
	return global.WithFields(fields...)
}

func Clone() *Logger {
	return global.Clone()
}

func SetLevel(level Level) {
	global.SetLevel(level)
}
func SetWriter(w Writer) {
	global.SetWriter(w)
}

func GetLevel() Level {
	return global.GetLevel()
}

func Enable(level Level) bool {
	return global.Enable(level)
}
