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

	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/registry"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"

	"github.com/imkuqin-zw/yggdrasil/pkg/application"
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

type restServiceDesc struct {
	ss     interface{}
	Prefix []string
}

type options struct {
	serviceDesc       map[*server.ServiceDesc]interface{}
	restServiceDesc   map[*server.RestServiceDesc]restServiceDesc
	restRawHandleDesc []*server.RestRawHandlerDesc
	server            server.Server
	governor          *governor.Server
	internalSvr       []application.InternalServer
	registry          registry.Registry
	shutdownTimeout   time.Duration
	startBeforeHook   []func() error
	stopBeforeHook    []func() error
	stopAfterHook     []func() error
}

func (opts *options) getAppOpts() []application.Option {
	return []application.Option{
		application.WithServer(opts.server),
		application.WithGovernor(opts.governor),
		application.WithRegistry(opts.registry),
		application.WithShutdownTimeout(opts.shutdownTimeout),
		application.WithBeforeStartHook(opts.startBeforeHook...),
		application.WithBeforeStopHook(opts.stopBeforeHook...),
		application.WithAfterStopHook(opts.stopAfterHook...),
		application.WithInternalServer(opts.internalSvr...),
	}
}

type Option func(*options) error

func WithBeforeStartHook(fns ...func() error) Option {
	return func(opts *options) error {
		opts.startBeforeHook = append(opts.startBeforeHook, fns...)
		return nil
	}
}

func WithBeforeStopHook(fns ...func() error) Option {
	return func(opts *options) error {
		opts.stopBeforeHook = append(opts.stopBeforeHook, fns...)
		return nil
	}
}

func WithAfterStopHook(fns ...func() error) Option {
	return func(opts *options) error {
		opts.stopAfterHook = append(opts.stopAfterHook, fns...)
		return nil
	}
}

func WithRegistry(registry registry.Registry) Option {
	return func(opts *options) error {
		opts.registry = registry
		return nil
	}
}

func WithGovernor(svr *governor.Server) Option {
	return func(opts *options) error {
		opts.governor = svr
		return nil
	}
}

func WithServiceDescMap(desc map[*server.ServiceDesc]interface{}) Option {
	return func(opts *options) error {
		for k, v := range desc {
			opts.serviceDesc[k] = v
		}
		return nil
	}
}

func WithServiceDesc(desc *server.ServiceDesc, impl interface{}) Option {
	return func(opts *options) error {
		opts.serviceDesc[desc] = impl
		return nil
	}
}

func WithRestServiceDesc(desc *server.RestServiceDesc, impl interface{}, prefix ...string) Option {
	return func(opts *options) error {
		opts.restServiceDesc[desc] = restServiceDesc{
			ss:     impl,
			Prefix: prefix,
		}
		return nil
	}
}

func WithRestRawHandleDesc(desc ...*server.RestRawHandlerDesc) Option {
	return func(opts *options) error {
		opts.restRawHandleDesc = append(opts.restRawHandleDesc, desc...)
		return nil
	}
}

func WithInternalServer(svr ...application.InternalServer) Option {
	return func(opts *options) error {
		opts.internalSvr = append(opts.internalSvr, svr...)
		return nil
	}
}
