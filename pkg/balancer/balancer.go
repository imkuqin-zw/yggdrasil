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

package balancer

import (
	"context"
	"fmt"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
)

var ErrNoAvailableInstance = fmt.Errorf("no available instance")

type RpcInfo struct {
	Ctx    context.Context
	Method string
}

type PickResult interface {
	Endpoint() resolver.Endpoint
	Report(err error)
}

type Picker interface {
	Next(RpcInfo) (PickResult, error)
}

type Balancer interface {
	GetPicker() Picker
	Update(config.Values)
	Close() error
	Name() string
}

type Builder func(serviceName string) Balancer

var (
	builder = map[string]Builder{}
	mu      sync.RWMutex
)

func GetBuilder(name string) (Builder, error) {
	mu.RLock()
	defer mu.RUnlock()
	f, ok := builder[name]
	if !ok {
		return nil, fmt.Errorf("not found balancer builder, name: %s", name)
	}
	return f, nil
}

func RegisterBuilder(name string, f Builder) {
	mu.Lock()
	defer mu.Unlock()
	builder[name] = f
}
