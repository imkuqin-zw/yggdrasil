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

package grpc

import (
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xtls"
	"google.golang.org/grpc"
)

type Config struct {
	Host         string
	Port         uint64
	Network      string
	TLS          *xtls.SSLConfig
	UnaryFilter  []string
	StreamFilter []string
	ServerOpts   []string

	serverOptions      []grpc.ServerOption
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
}

// WithServerOption inject grpcServer option to grpc grpcServer
// User should not inject interceptor option, which is recommend by WithStreamInterceptor
// and WithUnaryInterceptor
func (c *Config) WithServerOption(opts ...grpc.ServerOption) *Config {
	if c.serverOptions == nil {
		c.serverOptions = make([]grpc.ServerOption, 0, len(opts))
	}
	c.serverOptions = append(c.serverOptions, opts...)
	return c
}

// WithStreamInterceptor inject stream interceptors to grpcServer option
func (c *Config) WithStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) *Config {
	c.streamInterceptors = append(c.streamInterceptors, interceptors...)
	return c
}

// WithUnaryInterceptor inject unary interceptors to grpcServer option
func (c *Config) WithUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) *Config {
	c.unaryInterceptors = append(c.unaryInterceptors, interceptors...)
	return c
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) Build() *grpcServer {
	var err error
	c.Host, err = xnet.Extract(c.Host)
	if err != nil {
		log.Fatalf("fault to create grpc grpcServer, error: %s", err.Error())
	}
	if c.Network == "" {
		c.Network = "tcp"
	}

	c.UnaryFilter = xarray.RemoveReplaceStrings(append([]string{"error", "metadata", "log"}, c.UnaryFilter...))
	c.StreamFilter = xarray.RemoveReplaceStrings(append([]string{"error", "metadata", "log"}, c.StreamFilter...))

	for _, name := range c.UnaryFilter {
		f, ok := unaryInterceptor[name]
		if !ok {
			log.Warnf("not found grpc unary interceptor, name: %s", name)
		}
		c.WithUnaryInterceptor(f())
	}

	for _, name := range c.StreamFilter {
		f, ok := streamInterceptor[name]
		if !ok {
			log.Warnf("not found grpc stream interceptor, name: %s", name)
		}
		c.WithStreamInterceptor(f())
	}

	c.WithServerOption(
		grpc.WriteBufferSize(4096*2^10),
		grpc.ReadBufferSize(4096*2^10),
	)
	for _, name := range c.ServerOpts {
		f, ok := serverOptions[name]
		if !ok {
			log.Warnf("not found grpc server options, name: %s", name)
		}
		c.WithServerOption(f())
	}

	return newServer(c)
}
