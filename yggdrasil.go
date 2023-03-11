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
	"context"
	"errors"
	"sync/atomic"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/application"
	"github.com/imkuqin-zw/yggdrasil/pkg/client"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/registry"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
	"github.com/imkuqin-zw/yggdrasil/pkg/tracer"
	"go.opentelemetry.io/otel"
)

var (
	app         = application.New()
	initialized atomic.Bool
)

func NewServer() server.Server {
	return server.GetServer()
}

func NewClient(name string) client.Client {
	cli, err := client.NewClient(context.Background(), name)
	if err != nil {
		logger.FatalFiled("fault to new client", logger.String("name", name), logger.Err(err))
	}
	return cli
}

func Init(appName string, ops ...Option) error {
	if !initialized.CompareAndSwap(false, true) {
		return errors.New("yggdrasil had already init")
	}
	opts := &options{}
	initLogger()
	initInstanceInfo(appName)
	applyOpt(opts, ops...)
	initRegistry(opts)
	initTracer()
	initGovernor(opts)
	app.Init(opts.getAppOpts()...)
	return nil
}

func Start() error {
	if !initialized.Load() {
		return errors.New("please initialize the yggdrasil before starting")
	}
	return app.Run()
}

func Run(appName string, ops ...Option) error {
	if err := Init(appName, ops...); err != nil {
		return err
	}
	return app.Run()
}

func Stop() error {
	if err := app.Stop(); err != nil {
		logger.ErrorFiled("fault to stop yggdrasil application", logger.Err(err))
		return err
	}
	return nil
}

func initRegistry(opts *options) {
	name := config.GetString(config.KeyRegistry)
	if len(name) == 0 {
		return
	}
	f := registry.GetBuilder(name)
	if f == nil {
		logger.WarnFiled("not found registry", logger.String("name", name))
		return
	}
	_ = WithRegistry(f())(opts)
}

func initTracer() {
	if tracerName := config.GetString(config.KeyTracer); len(tracerName) > 0 {
		constructor := tracer.GetTracerProviderBuilder(tracerName)
		if constructor != nil {
			otel.SetTracerProvider(constructor(pkg.Name()))
		} else {
			logger.ErrorFiled("not found tracer provider", logger.String("name", tracerName))
		}
	}
}

func initInstanceInfo(appName string) {
	if err := config.Set(config.KeyAppName, appName); err != nil {
		logger.FatalFiled("fault to set application name", logger.Err(err))
	}
	pkg.InitInstanceInfo()
}

func initLogger() {
	logName := config.GetString(config.KeyLoggerName, "std")
	if logName == "std" {
		lv := config.GetBytes(config.KeyLoggerLevel, []byte("debug"))
		var level logger.Level
		if err := level.UnmarshalText(lv); err != nil {
			logger.FatalFiled("fault to unmarshal std logger level", logger.Err(err))
		}
		logger.SetLevel(level)
		if config.GetBool("stdLogger.openMsgFormat", false) {
			if lg, ok := logger.RawLogger().(*logger.StdLogger); ok {
				lg.OpenMsgFormat()
			}
		}
	} else {
		lg := logger.GetLogger(logName)
		logger.SetLogger(lg)
	}
	timeEncoder := config.GetString(config.KeyLoggerTimeEnc, "RFC3339")
	if err := logger.SetTimeEncoderByName(timeEncoder); err != nil {
		logger.FatalFiled("fault to set logger time encoder", logger.Err(err))
		return
	}
	durationEncoder := config.GetString(config.KeyLoggerDurEnc, "millis")
	if err := logger.SetDurationEncoderByName(durationEncoder); err != nil {
		logger.FatalFiled("fault to set logger duration encoder", logger.Err(err))
		return
	}
	logger.SetStackPrintState(config.GetBool(config.KeyLoggerStack, false))
}

func initGovernor(opts *options) {
	svr := governor.NewServer()
	_ = WithGovernor(svr)(opts)
}

func applyOpt(opts *options, ops ...Option) {
	for _, f := range ops {
		if err := f(opts); err != nil {
			logger.FatalFiled("fault to apply options", logger.Err(err))
		}
	}
}
