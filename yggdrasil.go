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
	xotel "github.com/imkuqin-zw/yggdrasil/pkg/otel"
	"github.com/imkuqin-zw/yggdrasil/pkg/registry"
	remotelg "github.com/imkuqin-zw/yggdrasil/pkg/remote/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"

	"go.opentelemetry.io/otel"
)

var (
	app         = application.New()
	appRunning  atomic.Bool
	initialized atomic.Bool
	opts        = &options{
		serviceDesc:     map[*server.ServiceDesc]interface{}{},
		restServiceDesc: map[*server.RestServiceDesc]restServiceDesc{},
	}
)

func NewClient(name string) client.Client {
	cli, err := client.NewClient(context.Background(), name)
	if err != nil {
		logger.FatalField("fault to new client", logger.String("name", name), logger.Err(err))
	}
	return cli
}

func Init(appName string, ops ...Option) {
	if !initialized.CompareAndSwap(false, true) {
		return
	}
	initLogger()
	initInstanceInfo(appName)
	applyOpt(opts, ops...)
	initGovernor(opts)
	initRegistry(opts)
	initTracer()
	initMeter()
	return
}

func Serve(ops ...Option) error {
	if !appRunning.CompareAndSwap(false, true) {
		return errors.New("application had already running")
	}
	if !initialized.Load() {
		return errors.New("please initialize the yggdrasil before starting")
	}
	applyOpt(opts, ops...)
	initServer(opts)
	app.Init(opts.getAppOpts()...)
	return app.Run()
}

func Run(appName string, ops ...Option) error {
	if !appRunning.CompareAndSwap(false, true) {
		return errors.New("application had already running")
	}
	Init(appName, ops...)
	initServer(opts)
	app.Init(opts.getAppOpts()...)
	return app.Run()
}

func Stop() error {
	if err := app.Stop(); err != nil {
		logger.ErrorField("fault to stop yggdrasil application", logger.Err(err))
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
		logger.WarnField("not found registry", logger.String("name", name))
		return
	}
	_ = WithRegistry(f())(opts)
}

func initTracer() {
	if tracerName := config.GetString(config.KeyTracer); len(tracerName) > 0 {
		constructor := xotel.GetTracerProviderBuilder(tracerName)
		if constructor != nil {
			otel.SetTracerProvider(constructor(pkg.Name()))
		} else {
			logger.ErrorField("not found tracer provider", logger.String("name", tracerName))
		}
	}
}

func initMeter() {
	if meterName := config.GetString(config.KeyMeter); len(meterName) > 0 {
		constructor := xotel.GetMeterProviderBuilder(meterName)
		if constructor != nil {
			otel.SetMeterProvider(constructor(pkg.Name()))
		} else {
			logger.ErrorField("not found meter provider", logger.String("name", meterName))
		}
	}
}

func initInstanceInfo(appName string) {
	if err := config.Set(config.KeyAppName, appName); err != nil {
		logger.FatalField("fault to set application name", logger.Err(err))
	}
	pkg.InitInstanceInfo()
}

func initLogger() {
	var lv logger.Level
	if err := lv.UnmarshalText(config.GetBytes(config.KeyLoggerLevel, []byte("debug"))); err != nil {
		logger.FatalField("fault to unmarshal global logger level", logger.Err(err))
	}
	logger.SetLevel(lv)
	timeEncoder := config.GetString(config.KeyLoggerTimeEnc, "RFC3339")
	if err := logger.SetTimeEncoderByName(timeEncoder); err != nil {
		logger.FatalField("fault to set global logger time encoder", logger.Err(err))
		return
	}
	durationEncoder := config.GetString(config.KeyLoggerDurEnc, "millis")
	if err := logger.SetDurationEncoderByName(durationEncoder); err != nil {
		logger.FatalField("fault to set global logger duration encoder", logger.Err(err))
		return
	}
	logger.SetStackPrintState(config.GetBool(config.KeyLoggerStack, false))
	// set global logger writer
	writer := config.GetString(config.KeyLoggerWriter, "golog")
	if writer == "golog" {
		writerCfg := &logger.WriterCfg{}
		if err := config.Get("golog").Scan(writerCfg); err != nil {
			logger.FatalField("fault to load golog writer config", logger.Err(err))
			return
		}
		logger.SetWriter(logger.NewWriter(writerCfg))
	} else {
		logger.SetWriter(logger.GetWriter(writer))
	}

	_ = config.AddWatcher(config.KeyLoggerLevel, func(event config.WatchEvent) {
		var lv logger.Level
		if err := lv.UnmarshalText(event.Value().Bytes([]byte("debug"))); err != nil {
			logger.ErrorField("fault to unmarshal global logger level", logger.Err(err))
			return
		}
		logger.SetLevel(lv)
	})
	// init remote logger
	var remoteLv logger.Level
	if err := remoteLv.UnmarshalText(config.GetBytes(config.KeyRemoteLgLevel, []byte("error"))); err != nil {
		logger.FatalField("fault to unmarshal remote logger level", logger.Err(err))
	}
	remotelg.Logger = logger.Clone().WithFields(logger.String("mod", "remote"))
	remotelg.Logger.SetLevel(remoteLv)
	_ = config.AddWatcher(config.KeyRemoteLgLevel, func(event config.WatchEvent) {
		var lv logger.Level
		if err := lv.UnmarshalText(event.Value().Bytes([]byte("debug"))); err != nil {
			logger.ErrorField("fault to unmarshal remote logger level", logger.Err(err))
			return
		}
		remotelg.Logger.SetLevel(lv)
	})
}

func initGovernor(opts *options) {
	svr := governor.NewServer()
	_ = WithGovernor(svr)(opts)
}

func initServer(opts *options) {
	var svr server.Server
	if len(opts.serviceDesc) > 0 {
		svr = server.NetServer()
		for k, v := range opts.serviceDesc {
			svr.RegisterService(k, v)
		}
	}
	if len(opts.restServiceDesc) > 0 {
		if svr == nil {
			svr = server.NetServer()
		}
		for k, v := range opts.restServiceDesc {
			svr.RegisterRestService(k, v.ss, v.Prefix...)
		}
	}
	opts.server = svr
}

func applyOpt(opts *options, ops ...Option) {
	for _, f := range ops {
		if err := f(opts); err != nil {
			logger.FatalField("fault to apply options", logger.Err(err))
		}
	}
}
