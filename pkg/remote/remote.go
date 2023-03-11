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

package remote

import (
	"context"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

var Logger = logger.Clone()

type Client interface {
	NewStream(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error)
	Close() error
	Scheme() string
}

type ClientBuilder func(context.Context, string, resolver.Endpoint) Client

type ServerInfo struct {
	Protocol string
	Address  string
	Attr     map[string]string
}

type Server interface {
	Serve() (<-chan error, error)
	Stop() error
	Info() ServerInfo
}

type MethodHandle func(ctx context.Context, sm string, ss stream.ServerStream) (interface{}, bool, error)

type ServerBuilder func(MethodHandle) (Server, error)

var (
	mu            sync.RWMutex
	clientBuilder = map[string]ClientBuilder{}
	serverBuilder = map[string]ServerBuilder{}
)

func RegisterClientBuilder(scheme string, builder ClientBuilder) {
	mu.Lock()
	defer mu.Unlock()
	clientBuilder[scheme] = builder
}

func GetClientBuilder(scheme string) ClientBuilder {
	mu.RLock()
	defer mu.RUnlock()
	builder, ok := clientBuilder[scheme]
	if !ok {
		return nil
	}
	return builder
}

func RegisterServerBuilder(scheme string, builder ServerBuilder) {
	mu.Lock()
	defer mu.Unlock()
	serverBuilder[scheme] = builder
}

func GetServerBuilder(scheme string) ServerBuilder {
	mu.RLock()
	defer mu.RUnlock()
	builder, ok := serverBuilder[scheme]
	if !ok {
		return nil
	}
	return builder
}
