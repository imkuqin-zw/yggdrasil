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
	"os"
)

type Helper struct {
	lg Logger
}

// Debug is debug level
func (h *Helper) Debug(args ...interface{}) {
	h.lg.Debug(args...)
}

// Info is info level
func (h *Helper) Info(args ...interface{}) {
	h.lg.Info(args...)
}

// Warn is warning level
func (h *Helper) Warn(args ...interface{}) {
	h.lg.Warn(args...)
}

// Error is reason level
func (h *Helper) Error(args ...interface{}) {
	h.lg.Error(args...)
}

// Error is fault level
func (h *Helper) Fatal(args ...interface{}) {
	h.lg.Fatal(args...)
	os.Exit(1)
}

// Debugf is format debug level
func (h *Helper) Debugf(fmt string, args ...interface{}) {
	h.lg.Debugf(fmt, args...)
}

// Infof is format info level
func (h *Helper) Infof(fmt string, args ...interface{}) {
	h.lg.Infof(fmt, args...)
}

// Warnf is format warning level
func (h *Helper) Warnf(fmt string, args ...interface{}) {
	h.lg.Warnf(fmt, args...)
}

// Errorf is format reason level
func (h *Helper) Errorf(fmt string, args ...interface{}) {
	h.lg.Errorf(fmt, args...)
}

// Fatalf is format fatal level
func (h *Helper) Fatalf(fmt string, args ...interface{}) {
	h.lg.Fatalf(fmt, args...)
	os.Exit(1)
}

func (h *Helper) GetRaw() Logger {
	return h.lg
}

func (h *Helper) DebugFiled(msg string, fields ...Field) {
	if h.lg.Enable(LvDebug) {
		buf, err := enc.Encode(fields)
		if err != nil {
			h.lg.Errorf("encode reason: %v", err)
			return
		}
		h.lg.Debugw(msg, "fields", json.RawMessage(buf.Bytes()))
	}
}

func (h *Helper) InfoFiled(msg string, fields ...Field) {
	if h.lg.Enable(LvInfo) {
		buf, err := enc.Encode(fields)
		if err != nil {
			h.lg.Errorf("encode reason: %v", err)
			return
		}
		h.lg.Infow(msg, "fields", json.RawMessage(buf.Bytes()))
	}
}

func (h *Helper) WarnFiled(msg string, fields ...Field) {
	if h.lg.Enable(LvWarn) {
		buf, err := enc.Encode(fields)
		if err != nil {
			h.lg.Errorf("encode reason: %v", err)
			return
		}
		h.lg.Warnw(msg, "fields", json.RawMessage(buf.Bytes()))
	}
}

func (h *Helper) ErrorFiled(msg string, fields ...Field) {
	if h.lg.Enable(LvError) {
		buf, err := enc.Encode(fields)
		if err != nil {
			h.lg.Errorf("encode reason: %v", err)
			return
		}
		h.lg.Errorw(msg, "fields", json.RawMessage(buf.Bytes()))
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

func (h *Helper) FatalFiled(msg string, fields ...Field) {
	if h.lg.Enable(LvFault) {
		buf, err := enc.Encode(fields)
		if err != nil {
			h.lg.Errorf("encode reason: %v", err)
			return
		}
		h.lg.Fatalw(msg, "fields", json.RawMessage(buf.Bytes()))
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

func (h *Helper) SetLevel(level Level) {
	h.lg.SetLevel(level)
}

func (h *Helper) GetLevel() Level {
	return h.lg.GetLevel()
}

func (h *Helper) Enable(level Level) bool {
	return h.lg.Enable(level)
}

func (h *Helper) Clone() *Helper {
	newHelper := *h
	newHelper.lg = h.lg.Clone()
	return &newHelper
}
