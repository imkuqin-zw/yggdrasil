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
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/application"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

type Stage application.Stage

const (
	StageMin Stage = iota
	// StageBeforeStop before app stop
	StageBeforeStop
	// StageAfterStop after app stop
	StageAfterStop

	StageMax
)

type options struct {
	servers         []types.Server
	registry        types.Registry
	shutdownTimeout time.Duration
	startBeforeHook []func() error
	stopBeforeHook  []func() error
	stopAfterHook   []func() error
}

func (opts *options) getAppOpts() []application.Option {
	return []application.Option{
		application.WithServers(opts.servers...),
		application.WithRegistry(opts.registry),
		application.WithShutdownTimeout(opts.shutdownTimeout),
		application.WithBeforeStartHook(opts.startBeforeHook...),
		application.WithBeforeStopHook(opts.stopBeforeHook...),
		application.WithAfterStopHook(opts.stopAfterHook...),
	}
}

type Option func(*options) error

func WithBeforeStartHook(fns ...func() error) Option {
	return func(app *options) error {
		app.startBeforeHook = append(app.startBeforeHook, fns...)
		return nil
	}
}

func WithBeforeStopHook(fns ...func() error) Option {
	return func(app *options) error {
		app.stopBeforeHook = append(app.stopBeforeHook, fns...)
		return nil
	}
}

func WithAfterStopHook(fns ...func() error) Option {
	return func(app *options) error {
		app.stopAfterHook = append(app.stopAfterHook, fns...)
		return nil
	}
}

func WithRegistry(registry types.Registry) Option {
	return func(Options *options) error {
		Options.registry = registry
		return nil
	}
}

func WithServers(servers ...types.Server) Option {
	return func(Options *options) error {
		Options.servers = append(Options.servers, servers...)
		return nil
	}
}
