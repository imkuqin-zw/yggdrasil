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

package yggdrasil

import (
	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/application"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/registry"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
	"github.com/imkuqin-zw/yggdrasil/pkg/trace"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.uber.org/atomic"
)

var app = application.New()
var initialized atomic.Bool

func Init(appName string, ops ...Option) error {
	opts := &options{}
	initLogger()
	initInstanceInfo(appName)
	applyOpt(opts, ops...)
	initServer(opts)
	initRegistry(opts)
	initTracer()
	app.Init(opts.getAppOpts()...)
	initialized.Store(true)
	return nil
}

func Start() error {
	if !initialized.Load() {
		return errors.New("please initialize the yggdrasil before starting")
	}
	return app.Run()
}

func Run(appName string, ops ...Option) error {
	opts := &options{}
	initLogger()
	initInstanceInfo(appName)
	applyOpt(opts, ops...)
	initServer(opts)
	initRegistry(opts)
	initTracer()
	app.Init(opts.getAppOpts()...)
	initialized.Store(true)
	return app.Run()
}

func Stop() error {
	if err := app.Stop(); err != nil {
		log.ErrorFiled("fault to stop yggdrasil application", log.Err(err))
		return err
	}
	return nil
}

func initServer(opts *options) {
	servers := make([]types.Server, 0)
	for _, f := range server.GetConstructors() {
		servers = append(servers, f())
	}
	if len(servers) > 0 {
		_ = WithServers(servers...)(opts)
	}
}

func initRegistry(opts *options) {
	registerName := config.GetString("yggdrasil.register")
	if len(registerName) == 0 {
		return
	}
	f := registry.GetConstructor(registerName)
	if f == nil {
		log.WarnFiled("not found registry", log.String("name", registerName))
		return
	}
	_ = WithRegistry(f())(opts)
}

func initTracer() {
	if tracerName := config.GetString("yggdrasil.tracer"); len(tracerName) > 0 {
		constructor := trace.GetConstructor(tracerName)
		if constructor != nil {
			otel.SetTracerProvider(constructor(pkg.Name()))
		} else {
			log.WarnFiled("not found tracer provider", log.String("name", tracerName))
		}
	}
}

func initInstanceInfo(appName string) {
	if err := config.Set("yggdrasil.application.name", appName); err != nil {
		log.FatalFiled("fault to set application name", log.Err(err))
	}
	pkg.InitInstanceInfo()
}

func initLogger() {
	logName := config.GetString("yggdrasil.logger.name", "std")
	if logName == "std" {
		lv := config.GetBytes("yggdrasil.logger.level", []byte("debug"))
		var level types.Level
		if err := level.UnmarshalText(lv); err != nil {
			log.FatalFiled("fault to unmarshal std logger level", log.Err(err))
		}
		log.SetLevel(level)
		if config.GetBool("stdLogger.openMsgFormat", false) {
			if lg, ok := log.GetRaw().(*log.StdLogger); ok {
				lg.OpenMsgFormat()
			}
		}
	} else {
		lg := log.GetLogger(logName)
		log.SetLogger(lg)
	}
	timeEncoder := config.GetString("yggdrasil.logger.timeEncoder", "RFC3339")
	if err := log.SetTimeEncoderByName(timeEncoder); err != nil {
		log.FatalFiled("fault to set log time encoder", log.Err(err))
		return
	}
	durationEncoder := config.GetString("yggdrasil.logger.durationEncoder", "millis")
	if err := log.SetDurationEncoderByName(durationEncoder); err != nil {
		log.FatalFiled("fault to set log duration encoder", log.Err(err))
		return
	}
	log.SetStackPrintState(config.GetBool("yggdrasil.logger.enablePrintStack", false))
}

func applyOpt(opts *options, ops ...Option) {
	for _, f := range ops {
		if err := f(opts); err != nil {
			log.FatalFiled("fault to apply options", log.Err(err))
		}
	}
}

func Endpoints() []types.ServerInfo {
	return app.Endpoints()
}
