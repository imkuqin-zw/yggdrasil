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
	"net/http"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/errors"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

func init() {
	RegisterUnaryInterceptor("log", LoggerUnaryClientInterceptor)
	RegisterStreamInterceptor("log", LoggerStreamClientInterceptor)
}

func getSlowThreshold(serverName string) time.Duration {
	return config.GetDuration(
		fmt.Sprintf("yggdrasil.client.%s.grpc.slowThreshold", serverName),
		config.GetDuration("yggdrasil.grpc.client.slowThreshold", time.Second),
	)
}

func LoggerUnaryClientInterceptor(serverName string) grpc.UnaryClientInterceptor {
	slowThreshold := getSlowThreshold(serverName)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		cost := time.Since(beg)
		fields := []log.Field{
			log.String("scheme", "grpc"),
			log.String("method", method),
			log.String("type", "unary"),
			log.Duration("cost", cost),
			log.Context(ctx),
		}
		if err != nil {
			e := errors.FromError(err)
			fields = append(fields, log.Int32("code", e.Code()), log.Err(err))
			if e.HttpCode() < http.StatusInternalServerError {
				log.WarnFiled("call", fields...)
			} else {
				log.ErrorFiled("call", fields...)
			}
			return err
		} else {
			fields = append(fields, log.Int32("code", int32(code.Code_OK)))
			if cost >= slowThreshold && log.Enable(types.LvWarn) {
				log.WarnFiled("call", fields...)
			} else {
				log.InfoFiled("call", fields...)
			}
		}
		return nil
	}
}

func LoggerStreamClientInterceptor(serverName string) grpc.StreamClientInterceptor {
	slowThreshold := getSlowThreshold(serverName)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		beg := time.Now()
		cs, err := streamer(ctx, desc, cc, method, opts...)
		cost := time.Since(beg)
		fields := []log.Field{
			log.String("scheme", "grpc"),
			log.String("method", method),
			log.String("type", "stream"),
			log.Duration("cost", cost),
		}
		if err != nil {
			e := errors.FromError(err)
			fields = append(fields, log.Int32("code", e.Code()), log.Err(err))
			if e.HttpCode() < http.StatusInternalServerError {
				log.WarnFiled("call", fields...)
			} else {
				log.ErrorFiled("call", fields...)
			}
			return cs, err
		} else {
			fields = append(fields, log.Int32("code", int32(code.Code_OK)))
			if cost >= slowThreshold && log.Enable(types.LvWarn) {
				log.WarnFiled("call", fields...)
			} else {
				log.InfoFiled("call", fields...)
			}
		}
		return cs, err
	}
}
