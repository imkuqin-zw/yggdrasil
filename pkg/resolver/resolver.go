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

package resolver

import (
	"fmt"
	"sync"
)

type Endpoint interface {
	GetAddress() string
	GetProtocol() string
	GetMetadata() map[string]interface{}
}

type Resolver interface {
	AddWatch(string) error
	DelWatch(string) error
	Close() error
	Name() string
}

var (
	resolver = map[string]Resolver{}
	builder  = map[string]func(name string) (Resolver, error){}
	mu       sync.RWMutex
)

func GetResolver(name string) (Resolver, error) {
	mu.RLocker()
	if r, ok := resolver[name]; ok {
		mu.RUnlock()
		return r, nil
	}
	mu.Lock()
	defer mu.Unlock()
	if r, ok := resolver[name]; ok {
		return r, nil
	}
	f, ok := builder[name]
	if !ok {
		return nil, fmt.Errorf("not found resolver builder, name: %s", name)
	}
	return f(name)
}

func DelResolver(name string) error {
	mu.Lock()
	defer mu.Unlock()
	r, ok := resolver[name]
	if !ok {
		return nil
	}
	return r.Close()
}

func RegisterBuilder(name string, f func(string) (Resolver, error)) {
	mu.Lock()
	defer mu.Unlock()
	builder[name] = f
}
