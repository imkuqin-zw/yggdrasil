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
	"context"
	"fmt"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getConfig(name string) *Config {
	c := &Config{}
	if err := config.Scan(fmt.Sprintf("yggdrasil.client.{%s}.grpc", name), c); err != nil {
		log.Fatalf("fault to get config, err: %s", err.Error())
		return nil
	}
	c.Name = name
	c.UnaryFilter = append(config.GetStringSlice("yggdrasil.grpc.unaryFilter"), c.UnaryFilter...)
	c.StreamFilter = append(config.GetStringSlice("yggdrasil.grpc.streamFilter"), c.StreamFilter...)
	return c
}

func DialByConfig(ctx context.Context, config *Config) *grpc.ClientConn {
	config.SetDefault()
	for _, name := range config.UnaryFilter {
		f, ok := unaryInterceptor[name]
		if !ok {
			log.WarnFiled("not found grpc unary interceptor", log.String("name", name))
			continue
		}
		config.WithUnaryInterceptor(f(config.Name))
	}
	for _, name := range config.StreamFilter {
		f, ok := streamInterceptor[name]
		if !ok {
			log.WarnFiled("not found grpc stream interceptor", log.String("name", name))
			continue
		}
		config.WithStreamInterceptor(f(config.Name))
	}
	config.WithDialOption(
		grpc.WithChainUnaryInterceptor(config.unaryInterceptors...),
		grpc.WithChainStreamInterceptor(config.streamInterceptors...),
	)

	if config.TLS != nil {
		tlsConfig, err := config.TLS.ServerTLSConfig()
		if err != nil {
			log.FatalFiled("fault to get tls config", log.Err(err))
		}
		config.WithDialOption(grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		config.WithDialOption(grpc.WithInsecure())
	}

	// 默认配置使用block
	if config.Block {
		if config.DialTimeout > time.Duration(0) {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, config.DialTimeout)
			defer cancel()
		}
		config.WithDialOption(grpc.WithBlock())
	}
	if config.KeepAlive != nil {
		config.WithDialOption(grpc.WithKeepaliveParams(*config.KeepAlive))
	}
	if config.Balancer != "" {
		config.WithDialOption(
			grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, config.Balancer)),
		)
	}
	cc, err := grpc.DialContext(ctx, config.Target, config.dialOptions...)
	if err != nil {
		if config.OnDialError == "panic" {
			log.FatalFiled("dial grpc", log.Err(err))
		} else {
			log.ErrorFiled("dial grpc", log.Err(err))
		}
	}
	return cc
}

func DialContext(ctx context.Context, ServerName string, opts ...grpc.DialOption) *grpc.ClientConn {
	return DialByConfig(ctx, getConfig(ServerName).WithDialOption(opts...))
}

func Dial(ServerName string, opts ...grpc.DialOption) *grpc.ClientConn {
	return DialByConfig(context.Background(), getConfig(ServerName).WithDialOption(opts...))
}
