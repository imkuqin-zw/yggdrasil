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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/errors"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
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
		if err != nil {
			e := errors.FromError(err)
			if e.HttpCode() < http.StatusInternalServerError {
				if log.Enable(types.LvWarn) {
					filed := map[string]interface{}{
						"scheme": "grpc",
						"method": method,
						"type":   "unary",
						"cost":   time.Since(beg).Seconds(),
						"err":    err.Error(),
						"code":   e.HttpCode(),
					}
					data, _ := json.Marshal(filed)
					log.Warnf("call\t%s", xstrings.Bytes2str(data))
				}
			} else {
				if log.Enable(types.LvError) {
					filed := map[string]interface{}{
						"scheme": "grpc",
						"method": method,
						"type":   "unary",
						"cost":   time.Since(beg).Seconds(),
						"err":    err.Error(),
						"code":   e.HttpCode(),
					}
					data, _ := json.Marshal(filed)
					log.Errorf("call\t%s", xstrings.Bytes2str(data))
				}
			}
			return err
		} else {
			if log.Enable(types.LvInfo) {
				cost := time.Since(beg)
				filed := map[string]interface{}{
					"scheme": "grpc",
					"method": method,
					"type":   "unary",
					"cost":   cost.Seconds(),
					"code":   http.StatusOK,
				}
				data, _ := json.Marshal(filed)
				if cost >= slowThreshold && log.Enable(types.LvWarn) {
					log.Warnf("call \t%s", xstrings.Bytes2str(data))
				} else {
					log.Infof("call\t%s", xstrings.Bytes2str(data))
				}
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
		if err != nil {
			e := errors.FromError(err)
			if e.HttpCode() < http.StatusInternalServerError {
				if log.Enable(types.LvWarn) {
					filed := map[string]interface{}{
						"scheme": "grpc",
						"method": method,
						"type":   "stream",
						"cost":   time.Since(beg),
						"err":    err.Error(),
						"code":   e.HttpCode(),
					}
					data, _ := json.Marshal(filed)
					log.Warnf("call\t%s", xstrings.Bytes2str(data))
				}
			} else {
				if log.Enable(types.LvError) {
					filed := map[string]interface{}{
						"scheme": "grpc",
						"method": method,
						"type":   "stream",
						"cost":   time.Since(beg),
						"err":    err.Error(),
						"code":   e.HttpCode(),
					}
					data, _ := json.Marshal(filed)
					log.Errorf("call\t%s", xstrings.Bytes2str(data))
				}
			}
		} else {
			if log.Enable(types.LvInfo) {
				cost := time.Since(beg)
				filed := map[string]interface{}{
					"scheme": "grpc",
					"method": method,
					"type":   "stream",
					"cost":   cost,
					"code":   http.StatusOK,
				}
				data, _ := json.Marshal(filed)
				if cost > slowThreshold && log.Enable(types.LvWarn) {
					log.Warnf("call\t%s", xstrings.Bytes2str(data))
				} else {
					log.Infof("call\t%s", xstrings.Bytes2str(data))
				}
			}
		}
		return cs, err
	}
}
