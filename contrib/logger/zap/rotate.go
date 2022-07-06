package zap

import (
	"path/filepath"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func newFileSyncer(config *FileConfig) zapcore.WriteSyncer {
	if config.Dir == "" {
		config.Dir = defaultFileDir
	}
	if config.Name == "" {
		config.Name = defaultFileName
	}
	if config.MaxSize == 0 {
		config.MaxSize = defaultFileMaxSize
	}
	if config.MaxBackup == 0 {
		config.MaxBackup = defaultFileMaxBackup
	}
	if config.MaxAge == 0 {
		config.MaxAge = defaultFileMaxAge
	}
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(config.Dir, config.Name),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackup,
		MaxAge:     config.MaxAge,
		LocalTime:  true,
		Compress:   false,
	})
}
