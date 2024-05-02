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

package logging

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
	"github.com/pkg/errors"
)

var name = "logger"

var (
	global *logging
	once   sync.Once
)

func initGlobalLogging() {
	once.Do(func() {
		cfg := Config{}
		if err := config.Get(fmt.Sprintf(config.KeyInterceptorCfg, name)).Scan(&cfg); err != nil {
			logger.ErrorField("fault to load logger config", logger.Err(errors.WithStack(err)))
		}
		global = &logging{cfg: &cfg}
	})
}

func init() {
	interceptor.RegisterUnaryClientIntBuilder(name, func(s string) interceptor.UnaryClientInterceptor {
		initGlobalLogging()
		return global.UnaryClientInterceptor
	})
	interceptor.RegisterStreamClientIntBuilder(name, func(s string) interceptor.StreamClientInterceptor {
		initGlobalLogging()
		return global.StreamClientInterceptor
	})
	interceptor.RegisterUnaryServerIntBuilder(name, func() interceptor.UnaryServerInterceptor {
		initGlobalLogging()
		return global.UnaryServerInterceptor
	})
	interceptor.RegisterStreamServerIntBuilder(name, func() interceptor.StreamServerInterceptor {
		initGlobalLogging()
		return global.UewStreamServerInterceptor
	})
}

type logging struct {
	cfg *Config
}

func (l *logging) UnaryServerInterceptor(ctx context.Context, req interface{}, info *interceptor.UnaryServerInfo, handler interceptor.UnaryHandler) (resp interface{}, err error) {
	startTime := time.Now()
	defer func() {
		var (
			st     = status.FromError(err)
			fields = make([]logger.Field, 0)
			event  = "normal"
			cost   = time.Since(startTime)
		)
		if l.cfg.SlowThreshold <= cost {
			event = "slow"
		}
		if rec := recover(); rec != nil {
			switch rec := rec.(type) {
			case error:
				err = rec
			default:
				err = fmt.Errorf("%v", rec)
			}
			st = status.FromError(err)
			stack := make([]byte, 4096)
			stack = stack[:runtime.Stack(stack, true)]
			fields = append(fields, logger.String("stack", xstrings.Bytes2str(stack)))
			event = "recover"
		}
		fields = append(fields,
			logger.Context(ctx),
			logger.String("type", "unary"),
			logger.String("method", info.FullMethod),
			logger.Float64("cost", float64(cost)/float64(time.Millisecond)),
			logger.Int32("code", st.Code()),
			logger.String("event", event))
		if l.cfg.PrintReqAndRes {
			fields = append(fields, logger.Reflect("req", req))
		}
		if err != nil {
			fields = append(fields, logger.Err(err))
			if st.HttpCode() >= http.StatusInternalServerError {
				logger.ErrorField("access", fields...)
			} else {
				logger.WarnField("access", fields...)
			}
		} else {
			if l.cfg.PrintReqAndRes {
				fields = append(fields, logger.Reflect("res", resp))
			}
			logger.InfoField("access", fields...)
		}
	}()
	return handler(ctx, req)
}

func (l *logging) UewStreamServerInterceptor(srv interface{}, ss stream.ServerStream, info *interceptor.StreamServerInfo, handler stream.StreamHandler) (err error) {
	startTime := time.Now()
	defer func() {
		var (
			st     = status.FromError(err)
			fields = make([]logger.Field, 0)
			event  = "normal"
			cost   = time.Since(startTime)
		)
		if rec := recover(); rec != nil {
			switch rec := rec.(type) {
			case error:
				err = rec
			default:
				err = fmt.Errorf("%v", rec)
			}
			st = status.FromError(err)
			stack := make([]byte, 4096)
			stack = stack[:runtime.Stack(stack, true)]
			fields = append(fields, logger.String("stack", xstrings.Bytes2str(stack)))
			event = "recover"
		}
		fields = append(fields,
			logger.Context(ss.Context()),
			logger.String("type", "stream"),
			logger.String("method", info.FullMethod),
			logger.Float64("cost", float64(cost)/float64(time.Millisecond)),
			logger.String("event", event),
			logger.Int32("code", st.Code()))
		if err != nil {
			fields = append(fields, logger.Err(err))
			if st.HttpCode() >= http.StatusInternalServerError {
				logger.ErrorField("access", fields...)
			} else {
				logger.WarnField("access", fields...)
			}
		} else {
			logger.InfoField("access", fields...)
		}
	}()
	return handler(srv, ss)
}

func (l *logging) UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, invoker interceptor.UnaryInvoker) (err error) {
	startTime := time.Now()
	defer func() {
		var (
			st     = status.FromError(err)
			fields = make([]logger.Field, 0)
			event  = "normal"
			cost   = time.Since(startTime)
		)
		if l.cfg.SlowThreshold <= cost {
			event = "slow"
		}
		if rec := recover(); rec != nil {
			switch rec := rec.(type) {
			case error:
				err = rec
			default:
				err = fmt.Errorf("%v", rec)
			}
			st = status.FromError(err)
			stack := make([]byte, 4096)
			stack = stack[:runtime.Stack(stack, true)]
			fields = append(fields, logger.String("stack", xstrings.Bytes2str(stack)))
			event = "recover"
		}
		fields = append(fields,
			logger.Context(ctx),
			logger.String("type", "unary"),
			logger.String("method", method),
			logger.Float64("cost", float64(cost)/float64(time.Millisecond)),
			logger.Int32("code", st.Code()),
			logger.String("event", event))
		if l.cfg.PrintReqAndRes {
			fields = append(fields, logger.Reflect("req", req))
		}
		if err != nil {
			fields = append(fields, logger.Err(err))
			if st.HttpCode() >= http.StatusInternalServerError {
				logger.ErrorField("access", fields...)
			} else {
				logger.WarnField("access", fields...)
			}
		} else {
			if l.cfg.PrintReqAndRes {
				fields = append(fields, logger.Reflect("res", reply))
			}
			if l.cfg.SlowThreshold <= cost {
				logger.WarnField("access", fields...)
			} else {
				logger.InfoField("access", fields...)
			}
		}
	}()
	err = invoker(ctx, method, req, reply)
	return

}

func (l *logging) StreamClientInterceptor(ctx context.Context, desc *stream.StreamDesc, method string, streamer interceptor.Streamer) (res stream.ClientStream, err error) {
	startTime := time.Now()
	defer func() {
		var (
			st     = status.FromError(err)
			fields = make([]logger.Field, 0)
			event  = "normal"
			cost   = time.Since(startTime)
		)
		if rec := recover(); rec != nil {
			switch rec := rec.(type) {
			case error:
				err = rec
			default:
				err = fmt.Errorf("%v", rec)
			}
			st = status.FromError(err)
			stack := make([]byte, 4096)
			stack = stack[:runtime.Stack(stack, true)]
			fields = append(fields, logger.String("stack", xstrings.Bytes2str(stack)))
			event = "recover"
		}
		fields = append(fields,
			logger.Context(ctx),
			logger.String("type", "stream"),
			logger.String("method", method),
			logger.Float64("cost", float64(cost)/float64(time.Millisecond)),
			logger.String("event", event),
			logger.Int32("code", st.Code()))
		if err != nil {
			fields = append(fields, logger.Err(err))
			if st.HttpCode() >= http.StatusInternalServerError {
				logger.ErrorField("access", fields...)
			} else {
				logger.WarnField("access", fields...)
			}
		} else {
			logger.InfoField("access", fields...)
		}
	}()
	return streamer(ctx, desc, method)
}
