package zap

import (
	"os"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	lg *zap.Logger
	*zap.SugaredLogger
	lv *zap.AtomicLevel
}

var _ types.Logger = (*Logger)(nil)

func (lg *Logger) SetLevel(lv types.Level) {
	switch lv {
	case types.LvDebug:
		lg.lv.SetLevel(zap.DebugLevel)
	case types.LvInfo:
		lg.lv.SetLevel(zap.InfoLevel)
	case types.LvWarn:
		lg.lv.SetLevel(zap.WarnLevel)
	case types.LvError:
		lg.lv.SetLevel(zap.ErrorLevel)
	case types.LvFault:
		lg.lv.SetLevel(zap.FatalLevel)
	}
}

func (lg *Logger) Enable(lv types.Level) bool {
	switch lv {
	case types.LvDebug:
		return lg.lv.Enabled(zap.DebugLevel)
	case types.LvInfo:
		return lg.lv.Enabled(zap.InfoLevel)
	case types.LvWarn:
		return lg.lv.Enabled(zap.WarnLevel)
	case types.LvError:
		return lg.lv.Enabled(zap.ErrorLevel)
	case types.LvFault:
		return lg.lv.Enabled(zap.FatalLevel)
	}
	return false
}

func (lg *Logger) GetLevel() types.Level {
	switch lg.lv.Level() {
	case zap.DebugLevel:
		return types.LvDebug
	case zap.InfoLevel:
		return types.LvInfo
	case zap.WarnLevel:
		return types.LvWarn
	case zap.ErrorLevel:
		return types.LvError
	case zap.FatalLevel:
		return types.LvFault
	}
	return types.LvDebug
}

func (lg *Logger) ZapLogger() *zap.Logger {
	return lg.lg
}

func NewLogger(config *Config) *Logger {
	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.PanicLevel))
	if config.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}

	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}
	cores := make([]zapcore.Core, 0, 1)
	isErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && lv.Level() <= zapcore.ErrorLevel
	})
	isNotErr := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lv.Level() <= lvl
	})

	if config.Console.Enable {
		var wsOut, wsErr = zapcore.Lock(os.Stdout), zapcore.Lock(os.Stderr)
		var encoder = zapcore.NewConsoleEncoder(*config.Console.Encoder)
		cores = append(cores,
			zapcore.NewCore(encoder, wsErr, isErr),
			zapcore.NewCore(encoder, wsOut, isNotErr),
		)
	}
	if config.File.Enable {
		ws := zapcore.AddSync(newFileSyncer(&config.File.FileConfig))
		encoder := zapcore.NewJSONEncoder(*config.File.Encoder)
		cores = append(cores, zapcore.NewCore(encoder, ws, lv))
	}

	lg := zap.New(zapcore.NewTee(cores...), zapOptions...)
	return &Logger{
		lg:            lg,
		SugaredLogger: lg.Sugar(),
		lv:            &lv,
	}
}
