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

package defers

import (
	"sync"
)

var globalDefers = &Defer{fns: make([]func() error, 0)}

type Defer struct {
	sync.Mutex
	fns []func() error
}

func NewDefer() *Defer {
	return &Defer{
		fns: make([]func() error, 0),
	}
}

func (d *Defer) Register(fns ...func() error) {
	d.Lock()
	defer d.Unlock()
	d.fns = append(d.fns, fns...)
}

func (d *Defer) Done() {
	d.Lock()
	defer d.Unlock()
	for i := len(d.fns) - 1; i >= 0; i-- {
		_ = d.fns[i]()
	}
}

// Register 注册一个defer函数
func Register(fns ...func() error) {
	globalDefers.Register(fns...)
}

// Clean 清除
func Done() {
	globalDefers.Done()
}
