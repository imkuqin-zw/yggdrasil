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
				TimeKey:       "ts",
				LevelKey:      "lv",
				NameKey:       "Logger",
				CallerKey:     "caller",
				MessageKey:    "msg",
				StacktraceKey: "stack",
				LineEnding:    zapcore.DefaultLineEnding,
				EncodeLevel:   zapcore.LowercaseLevelEncoder,
				EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
					encoder.AppendInt64(t.Unix())
				},
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}
		}
	}

	if config.Console.Enable {
		if config.Console.Encoder == nil {
			config.Console.Encoder = &zapcore.EncoderConfig{
				TimeKey:       "ts",
				LevelKey:      "lv",
				NameKey:       "Logger",
				CallerKey:     "caller",
				MessageKey:    "msg",
				StacktraceKey: "stack",
				LineEnding:    zapcore.DefaultLineEnding,
				EncodeLevel:   consoleEncodeLevel,
				EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
					enc.AppendString(t.Format("2006-01-02 15:04:05"))
				},
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}
		}
	}
	return newLogger(config)
}