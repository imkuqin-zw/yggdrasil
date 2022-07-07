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
	initServer(opts)
	initRegistry(opts)
	applyOpt(opts, ops...)
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
	initServer(opts)
	initRegistry(opts)
	applyOpt(opts, ops...)
	initTracer()
	app.Init(opts.getAppOpts()...)
	initialized.Store(true)
	return app.Run()
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
		log.Warnf("not found registry, name: %s", registerName)
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
			log.Warnf("not found tracer provider, name: %s", tracerName)
		}
	}
}

func initInstanceInfo(appName string) {
	if err := config.Set("yggdrasil.application.name", appName); err != nil {
		log.Fatalf("fault to set application name, err: %s", err.Error())
	}
	pkg.InitInstanceInfo()
}

func initLogger() {
	logName := config.GetString("yggdrasil.logger.name", "std")
	if logName == "std" {
		lv := config.GetBytes("yggdrasil.logger.level", []byte("debug"))
		var level types.Level
		if err := level.UnmarshalText(lv); err != nil {
			log.Fatalf("fault to unmarshal std logger level, err: %s", err.Error())
		}
		log.SetLevel(level)
	} else {
		lg := log.GetLogger(logName)
		log.SetLogger(lg)
	}
}

func applyOpt(opts *options, ops ...Option) {
	for _, f := range ops {
		if err := f(opts); err != nil {
			log.Fatalf("fault to apply options, err: %s", err.Error())
		}
	}
}