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

package xgo

import (
	"context"
	"runtime/debug"

	"github.com/imkuqin-zw/yggdrasil/pkg/log"
)

func Go(f func(), recoverHandle func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("goroutine panic: %v\n%s\n", r, string(debug.Stack()))
				if recoverHandle != nil {
					go func() {
						defer func() {
							if p := recover(); p != nil {
								log.Errorf("goroutine panic: %v\n%s\n", p, string(debug.Stack()))
							}
						}()
						recoverHandle(r)
					}()
				}
			}
		}()
		f()
	}()
}

func GoWithCtx(ctx context.Context, f func(ctx context.Context), recoverHandle func(context.Context, interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("goroutine panic: %v\n%s\n", r, string(debug.Stack()))
				if recoverHandle != nil {
					go func() {
						defer func() {
							if p := recover(); p != nil {
								log.Errorf("goroutine panic: %v\n%s\n", p, string(debug.Stack()))
							}
						}()
						recoverHandle(ctx, r)
					}()
				}
			}
		}()
		f(ctx)
	}()
}
