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

package logger

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
	"github.com/pkg/errors"
)

func init() {
	interceptor.RegisterUnaryClientIntBuilder(name, newUnaryClientInterceptor)
	interceptor.RegisterStreamClientIntBuilder(name, newStreamClientInterceptor)
	interceptor.RegisterUnaryServerIntBuilder(name, newUnaryServerInterceptor)
	interceptor.RegisterStreamServerIntBuilder(name, newStreamServerInterceptor)
}

type Config struct {
	SlowThreshold  time.Duration `default:"1s"`
	PrintReqAndRes bool
}

var name = "logger"

func newUnaryServerInterceptor() interceptor.UnaryServerInterceptor {
	cfg := Config{}
	if err := config.Get(fmt.Sprintf(config.KeyInterceptorCfg, name)).Scan(&cfg); err != nil {
		logger.ErrorFiled("fault to load logger config", logger.Err(errors.WithStack(err)))
	}
	return func(ctx context.Context, req interface{}, info *interceptor.UnaryServerInfo, handler interceptor.UnaryHandler) (resp interface{}, err error) {
		var (
			startTime = time.Now()
		)
		defer func() {
			var (
				st     = status.FromError(err)
				fields = make([]logger.Field, 0)
				event  = "normal"
				cost   = time.Since(startTime)
			)
			if cfg.SlowThreshold <= cost {
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
				logger.String("type", "unary"),
				logger.String("method", info.FullMethod),
				logger.Duration("cost", cost),
				logger.Int32("code", st.Code()),
				logger.String("event", event))
			if cfg.PrintReqAndRes {
				fields = append(fields, logger.Reflect("req", req))
			}
			if err != nil {
				fields = append(fields, logger.Err(err))
				if st.HttpCode() >= http.StatusInternalServerError {
					logger.ErrorFiled("access", fields...)
				} else {
					logger.WarnFiled("access", fields...)
				}
			} else {
				if cfg.PrintReqAndRes {
					fields = append(fields, logger.Reflect("res", resp))
				}
				logger.InfoFiled("access", fields...)
			}
		}()
		return handler(ctx, req)
	}
}

func newStreamServerInterceptor() interceptor.StreamServerInterceptor {
	return func(srv interface{}, stream stream.ServerStream, info *interceptor.StreamServerInfo, handler stream.StreamHandler) (err error) {
		var (
			startTime = time.Now()
		)
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
				logger.String("type", "stream"),
				logger.String("method", info.FullMethod),
				logger.Duration("cost", cost),
				logger.String("event", event),
				logger.Int32("code", st.Code()))
			if err != nil {
				fields = append(fields, logger.Err(err))
				if st.HttpCode() >= http.StatusInternalServerError {
					logger.ErrorFiled("access", fields...)
				} else {
					logger.WarnFiled("access", fields...)
				}
			} else {
				logger.InfoFiled("access", fields...)
			}
		}()
		return handler(srv, stream)
	}
}

func newUnaryClientInterceptor(serverName string) interceptor.UnaryClientInterceptor {
	cfg := &Config{}
	if err := config.GetMulti(
		fmt.Sprintf(config.KeyInterceptorCfg, name),
		fmt.Sprintf(config.KeyClientIntCfg, serverName, name),
	).Scan(cfg); err != nil {
		logger.ErrorFiled("fault to load logger config", logger.Err(errors.WithStack(err)))
	}
	return func(ctx context.Context, method string, req, reply interface{}, invoker interceptor.UnaryInvoker) (err error) {
		startTime := time.Now()
		defer func() {
			var (
				st     = status.FromError(err)
				fields = make([]logger.Field, 0)
				event  = "normal"
				cost   = time.Since(startTime)
			)
			if cfg.SlowThreshold <= cost {
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
				logger.String("type", "unary"),
				logger.String("method", method),
				logger.Duration("cost", cost),
				logger.Int32("code", st.Code()),
				logger.String("event", event))
			if cfg.PrintReqAndRes {
				fields = append(fields, logger.Reflect("req", req))
			}
			if err != nil {
				fields = append(fields, logger.Err(err))
				if st.HttpCode() >= http.StatusInternalServerError {
					logger.ErrorFiled("access", fields...)
				} else {
					logger.WarnFiled("access", fields...)
				}
			} else {
				if cfg.PrintReqAndRes {
					fields = append(fields, logger.Reflect("res", reply))
				}
				if cfg.SlowThreshold <= cost {
					logger.WarnFiled("access", fields...)
				} else {
					logger.InfoFiled("access", fields...)
				}
			}

		}()
		err = invoker(ctx, method, req, reply)
		return
	}
}

func newStreamClientInterceptor(string) interceptor.StreamClientInterceptor {
	return func(ctx context.Context, desc *stream.StreamDesc, method string, streamer interceptor.Streamer) (res stream.ClientStream, err error) {
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
				logger.String("type", "stream"),
				logger.String("method", method),
				logger.Duration("cost", cost),
				logger.String("event", event),
				logger.Int32("code", st.Code()))
			if err != nil {
				fields = append(fields, logger.Err(err))
				if st.HttpCode() >= http.StatusInternalServerError {
					logger.ErrorFiled("access", fields...)
				} else {
					logger.WarnFiled("access", fields...)
				}
			} else {
				logger.InfoFiled("access", fields...)
			}
		}()
		return streamer(ctx, desc, method)
	}
}
