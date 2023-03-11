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
	"log"

	"errors"
)

var (
	helper     *Helper
	enc        Encoder
	printStack bool
)

func init() {
	lg := &StdLogger{level: LvDebug, lg: log.Default(), kvsMsgFormat: "%-8s%s "}
	enc = &jsonEncoder{
		EncodeTime:     RFC3339TimeEncoder,
		EncodeDuration: MillisDurationEncoder,
		spaced:         false,
		buf:            Get(),
	}
	printStack = true
	helper = &Helper{lg}
}

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *Level")

const (
	LvDebug Level = iota - 1
	LvInfo
	LvWarn
	LvError
	LvFault
)

type LoggerConstructor func() Logger

func SetLogger(logger Logger) {
	helper.lg = logger
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

var loggerConstructors = make(map[string]LoggerConstructor)

func RegisterConstructor(name string, f LoggerConstructor) {
	loggerConstructors[name] = f
}

func GetConstructor(name string) LoggerConstructor {
	f, _ := loggerConstructors[name]
	return f
}

func GetLogger(name string) Logger {
	f := GetConstructor(name)
	if f == nil {
		log.Fatalf("unknown logger constructor, name: %s", name)
	}
	return f()
}

func Debug(args ...interface{}) {
	helper.Debug(args...)
}

func Info(args ...interface{}) {
	helper.Info(args...)
}

func Warn(args ...interface{}) {
	helper.Warn(args...)
}

func Error(args ...interface{}) {
	helper.Error(args...)
}

func Fatal(args ...interface{}) {
	helper.Fatal(args...)
}

func Debugf(fmt string, args ...interface{}) {
	helper.Debugf(fmt, args)
}

func Infof(fmt string, args ...interface{}) {
	helper.Infof(fmt, args)
}

func Warnf(fmt string, args ...interface{}) {
	helper.Warnf(fmt, args)
}

func Errorf(fmt string, args ...interface{}) {
	helper.Errorf(fmt, args)
}

func Fatalf(fmt string, args ...interface{}) {
	helper.Fatalf(fmt, args)
}

func DebugFiled(msg string, fields ...Field) {
	helper.DebugFiled(msg, fields...)
}

func InfoFiled(msg string, fields ...Field) {
	helper.InfoFiled(msg, fields...)
}

func WarnFiled(msg string, fields ...Field) {
	helper.WarnFiled(msg, fields...)
}

func ErrorFiled(msg string, fields ...Field) {
	helper.ErrorFiled(msg, fields...)
}

func FatalFiled(msg string, fields ...Field) {
	helper.FatalFiled(msg, fields...)
}

func Clone() *Helper {
	return helper.Clone()
}

func RawLogger() Logger {
	return helper.lg
}

func SetLevel(level Level) {
	helper.SetLevel(level)
}

func GetLevel() Level {
	return helper.GetLevel()
}

func Enable(level Level) bool {
	return helper.Enable(level)
}
