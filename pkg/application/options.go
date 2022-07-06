package application

import (
	"fmt"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

type Option func(*Application) error

func WithHook(stage Stage, fns ...func() error) Option {
	return func(app *Application) error {
		hooks, ok := app.hooks[stage]
		if ok {
			hooks.Register(fns...)
			return nil
		}
		return fmt.Errorf("hook stage not found")
	}
}

func WithBeforeStopHook(fns ...func() error) Option {
	return WithHook(StageBeforeStop, fns...)
}

func WithAfterStopHook(fns ...func() error) Option {
	return WithHook(StageAfterStop, fns...)
}

func WithRegistry(registry types.Registry) Option {
	return func(application *Application) error {
		application.registry = registry
		return nil
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(application *Application) error {
		application.shutdownTimeout = timeout
		return nil
	}
}

func WithServers(servers ...types.Server) Option {
	return func(application *Application) error {
		application.servers = append(application.servers, servers...)
		return nil
	}
}
