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
	"github.com/imkuqin-zw/yggdrasil/pkg/stats"
)

type ClientBuilder func(context.Context, string, resolver.Endpoint, stats.Handler) Client

type ServerInfo struct {
	Protocol string
	Address  string
	Attr     map[string]string
}

type ServerBuilder func(handle MethodHandle) (Server, error)

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
