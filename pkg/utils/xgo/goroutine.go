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
