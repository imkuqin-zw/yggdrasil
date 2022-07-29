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
