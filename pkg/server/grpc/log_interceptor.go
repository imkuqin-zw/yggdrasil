package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
	"google.golang.org/grpc"
)

func init() {
	RegisterUnaryInterceptor("log", LogUnaryServerInterceptor)
	RegisterStreamInterceptor("log", LogStreamServerInterceptor)
}

// StreamServerInterceptor returns a new streaming grpcServer interceptor that adds zap.Logger to the context.
func LogStreamServerInterceptor() grpc.StreamServerInterceptor {
	slowThreshold := config.GetDuration("yggdrasil.server.grpc.slowThreshold", time.Second)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		var (
			startTime = time.Now()
			stack     = ""
		)
		defer func() {
			cost := time.Since(startTime)
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}
				stackArr := make([]byte, 4096)
				stack = xstrings.Bytes2str(stackArr[:runtime.Stack(stackArr, true)])
			}
			if err != nil {
				if log.Enable(types.LvError) {
					filed := map[string]interface{}{
						"scheme": "grpc",
						"type":   "stream",
						"method": info.FullMethod,
						"cost":   cost.Seconds(),
						"err":    err.Error(),
					}
					if len(stack) != 0 {
						filed["stack"] = stack
					}
					data, _ := json.Marshal(filed)
					log.Errorf("access\t%s", xstrings.Bytes2str(data))
				}
				return
			}
			if log.Enable(types.LvInfo) {
				filed := map[string]interface{}{
					"scheme": "grpc",
					"type":   "stream",
					"method": info.FullMethod,
					"cost":   cost.Seconds(),
				}
				data, _ := json.Marshal(filed)
				if log.Enable(types.LvWarn) && slowThreshold <= cost {
					log.Warnf("access\t%s", xstrings.Bytes2str(data))
				} else {
					log.Infof("access\t%s", xstrings.Bytes2str(data))
				}
			}
		}()
		return handler(srv, stream)
	}
}

// UnaryServerInterceptor returns a new unary grpcServer interceptors that adds zap.Logger to the context.
func LogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	slowThreshold := config.GetDuration("yggdrasil.server.grpc.slowThreshold", time.Second)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			startTime = time.Now()
			stack     = ""
		)
		defer func() {
			cost := time.Since(startTime)
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}
				stackArr := make([]byte, 4096)
				stack = xstrings.Bytes2str(stackArr[:runtime.Stack(stackArr, true)])
			}
			if err != nil {
				if log.Enable(types.LvError) {
					filed := map[string]interface{}{
						"scheme": "grpc",
						"type":   "unary",
						"method": info.FullMethod,
						"cost":   cost.Seconds(),
						"err":    err.Error(),
					}
					if len(stack) != 0 {
						filed["stack"] = stack
					}
					data, _ := json.Marshal(filed)
					log.Errorf("access\t%s", xstrings.Bytes2str(data))
				}
				return
			}
			if log.Enable(types.LvInfo) {
				filed := map[string]interface{}{
					"scheme": "grpc",
					"type":   "unary",
					"method": info.FullMethod,
					"cost":   cost.Seconds(),
				}
				data, _ := json.Marshal(filed)
				if log.Enable(types.LvWarn) && slowThreshold <= cost {
					log.Warnf("access\t%s", xstrings.Bytes2str(data))
				} else {
					log.Infof("access\t%s", xstrings.Bytes2str(data))
				}

			}
		}()
		return handler(ctx, req)
	}
}
