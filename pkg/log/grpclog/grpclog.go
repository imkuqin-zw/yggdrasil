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

package grpclog

import (
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"google.golang.org/grpc/grpclog"
)

var once sync.Once

func SetLogger() {
	once.Do(func() {
		grpclog.SetLoggerV2(&loggerWrapper{})
	})
}

// loggerWrapper wraps x*log.Logger into a LoggerV2.
type loggerWrapper struct {
}

// Info logs to INFO log
func (l *loggerWrapper) Info(args ...interface{}) {
	log.Info(args...)
}

// Infoln logs to INFO log
func (l *loggerWrapper) Infoln(args ...interface{}) {
	log.Info(args...)
}

// Infof logs to INFO log
func (l *loggerWrapper) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warning(args ...interface{}) {
	log.Warn(args...)
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warningln(args ...interface{}) {
	log.Warn(args...)
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warningf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Error logs to ERROR log
func (l *loggerWrapper) Error(args ...interface{}) {
	log.Error(args...)
}

// Errorn logs to ERROR log
func (l *loggerWrapper) Errorln(args ...interface{}) {
	log.Error(args...)
}

// Errorf logs to ERROR log
func (l *loggerWrapper) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatal logs to ERROR log
func (l *loggerWrapper) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalln logs to ERROR log
func (l *loggerWrapper) Fatalln(args ...interface{}) {
	log.Fatal(args...)
}

// Error logs to ERROR log
func (l *loggerWrapper) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// v returns true for all verbose level.
func (l *loggerWrapper) V(v int) bool {
	return true
}
