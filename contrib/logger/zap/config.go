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

package zap

import (
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xcolor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultFileDir = "."

	defaultFileName = "out.log"

	defaultFileMaxSize = 500

	defaultFileMaxAge = 1

	defaultFileMaxBackup = 10
)

const (
	// defaultBufferSize sizes the buffer associated with each WriterSync.
	defaultBufferSize = 256 * 1024

	// defaultFlushInterval means the default flush interval
	defaultFlushInterval = 30 * time.Second
)

// FileConfig
type FileConfig struct {
	Dir       string
	Name      string
	MaxSize   int
	MaxBackup int
	MaxAge    int
}

// BufferConfig
type BufferConfig struct {
	BufferSize    int
	FlushInterval time.Duration
}

// Config
type Config struct {
	Level     string
	AddCaller bool
	File      struct {
		Enable bool
		FileConfig
		Encoder *zapcore.EncoderConfig
	}
	Console struct {
		Enable  bool
		Encoder *zapcore.EncoderConfig
	}
}

// DebugEncodeLevel ...
func consoleEncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorize = xcolor.Red
	switch lv {
	case zapcore.DebugLevel:
		colorize = xcolor.Blue
	case zapcore.InfoLevel:
		colorize = xcolor.Green
	case zapcore.WarnLevel:
		colorize = xcolor.Yellow
	case zapcore.ErrorLevel, zap.PanicLevel, zap.DPanicLevel, zap.FatalLevel:
		colorize = xcolor.Red
	default:
	}
	enc.AppendString(colorize(lv.CapitalString()))
}

func (config *Config) Build() *Logger {
	if config.File.Enable {
		if config.File.Encoder == nil {
			config.File.Encoder = &zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "lv",
				NameKey:        "Logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stack",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.EpochMillisTimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}
		}
	}

	if config.Console.Enable {
		if config.Console.Encoder == nil {
			config.Console.Encoder = &zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "lv",
				NameKey:        "Logger",
				CallerKey:      "caller",
				MessageKey:     "msg",
				StacktraceKey:  "stack",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    consoleEncodeLevel,
				EncodeTime:     zapcore.RFC3339TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}
		}
	}
	return newLogger(config)
}
