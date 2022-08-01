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

package log

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

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

func DebugFiled(msg string, fields ...Field) {
	if Enable(types.LvDebug) {
		buf, err := enc.Encode(fields)
		if err != nil {
			lg.Errorf("encode error: %v", err)
			return
		}
		lg.Debugw(msg, "fields", json.RawMessage(buf.Bytes()))
	}
}

func InfoFiled(msg string, fields ...Field) {
	if Enable(types.LvInfo) {
		buf, err := enc.Encode(fields)
		if err != nil {
			lg.Errorf("encode error: %v", err)
			return
		}
		lg.Infow(msg, "fields", json.RawMessage(buf.Bytes()))
	}
}

func WarnFiled(msg string, fields ...Field) {
	if Enable(types.LvWarn) {
		buf, err := enc.Encode(fields)
		if err != nil {
			lg.Errorf("encode error: %v", err)
			return
		}
		lg.Warnw(msg, "fields", json.RawMessage(buf.Bytes()))
	}
}

func ErrorFiled(msg string, fields ...Field) {
	if Enable(types.LvError) {
		buf, err := enc.Encode(fields)
		if err != nil {
			lg.Errorf("encode error: %v", err)
			return
		}
		lg.Errorw(msg, "fields", json.RawMessage(buf.Bytes()))
		if printStack {
			for _, item := range fields {
				if item.Type == ErrorType {
					if item, ok := item.Interface.(fmt.Formatter); ok {
						_, _ = fmt.Fprintf(os.Stderr, "%+v\n", item)
					}
				}
			}
		}
	}
}

func FatalFiled(msg string, fields ...Field) {
	if Enable(types.LvFault) {
		buf, err := enc.Encode(fields)
		if err != nil {
			lg.Errorf("encode error: %v", err)
			return
		}
		lg.Fatalw(msg, "fields", json.RawMessage(buf.Bytes()))
		if printStack {
			for _, item := range fields {
				if item.Type == ErrorType {
					if item, ok := item.Interface.(fmt.Formatter); ok {
						_, _ = fmt.Fprintf(os.Stderr, "%+v\n", item)
					}
				}
			}
		}
	}
	os.Exit(1)
}
