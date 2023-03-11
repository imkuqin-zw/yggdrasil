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

package application

import (
	"fmt"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/registry"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
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

func WithBeforeStartHook(fns ...func() error) Option {
	return WithHook(StageBeforeStart, fns...)
}

func WithAfterStopHook(fns ...func() error) Option {
	return WithHook(StageAfterStop, fns...)
}

func WithRegistry(registry registry.Registry) Option {
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

func WithServer(server server.Server) Option {
	return func(application *Application) error {
		application.server = server
		return nil
	}
}

func WithGovernor(svr *governor.Server) Option {
	return func(application *Application) error {
		application.governor = svr
		return nil
	}
}
