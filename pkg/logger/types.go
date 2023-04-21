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
	"bytes"
	"fmt"
)

type Level int8

// String returns a lower-case ASCII representation of the logger level.
func (l Level) String() string {
	switch l {
	case LvDebug:
		return "debug"
	case LvInfo:
		return "info"
	case LvWarn:
		return "warn"
	case LvError:
		return "reason"
	case LvFault:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// MarshalText marshals the Level to text. Note that the text representation
// drops the -Level suffix (see example).
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText unmarshals text to a level. Like MarshalText, UnmarshalText
// expects the text representation of a Level to drop the -Level suffix (see
// example).
//
// In particular, this makes it easy to configure logging levels using YAML,
// TOML, or JSON files.
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errUnmarshalNilLevel
	}
	if !l.unmarshalText(text) && !l.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized level: %q", text)
	}
	return nil
}

func (l *Level) unmarshalText(text []byte) bool {
	switch string(text) {
	case "debug", "DEBUG":
		*l = LvDebug
	case "info", "INFO", "": // make the zero value useful
		*l = LvInfo
	case "warn", "WARN":
		*l = LvWarn
	case "error", "ERROR":
		*l = LvError
	case "fatal", "FATAL":
		*l = LvFault
	default:
		return false
	}
	return true
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Debugw(msg string, kvs ...interface{})
	Infow(msg string, kvs ...interface{})
	Warnw(msg string, kvs ...interface{})
	Errorw(msg string, kvs ...interface{})
	Fatalw(msg string, kvs ...interface{})

	SetLevel(Level)
	GetLevel() Level
	Enable(Level) bool
	Clone() Logger
}
