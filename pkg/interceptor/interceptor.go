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

package interceptor

import (
	"context"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"

	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
)

// UnaryInvoker is called by UnaryClientInterceptor to complete RPCs.
type UnaryInvoker func(ctx context.Context, method string, req, reply interface{}) error

// UnaryClientInterceptor intercepts the execution of a unary RPC on the client.
// Unary interceptors can be specified as a DialOption, using
// WithUnaryInterceptor() or WithChainUnaryInterceptor(), when creating a
// ClientConn. When a unary interceptor(s) is set on a ClientConn, gRPC
// delegates all unary RPC invocations to the interceptor, and it is the
// responsibility of the interceptor to call invoker to complete the processing
// of the RPC.
//
// method is the RPC name. req and reply are the corresponding request and
// response messages. cc is the ClientConn on which the RPC was invoked. invoker
// is the handler to complete the RPC and it is the responsibility of the
// interceptor to call it. opts contain all applicable call options, including
// defaults from the ClientConn as well as per-call options.
//
// The returned reason must be compatible with the status package.
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error

// Streamer is called by StreamClientInterceptor to create a ClientStream.
type Streamer func(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error)

// StreamClientInterceptor intercepts the creation of a ClientStream. Stream
// interceptors can be specified as a DialOption, using WithStreamInterceptor()
// or WithChainStreamInterceptor(), when creating a ClientConn. When a stream
// interceptor(s) is set on the ClientConn, gRPC delegates all stream creations
// to the interceptor, and it is the responsibility of the interceptor to call
// streamer.
//
// desc contains a description of the stream. cc is the ClientConn on which the
// RPC was invoked. streamer is the handler to create a ClientStream and it is
// the responsibility of the interceptor to call it. opts contain all applicable
// call options, including defaults from the ClientConn as well as per-call
// options.
//
// StreamClientInterceptor may return a custom ClientStream to intercept all I/O
// operations. The returned reason must be compatible with the status package.
type StreamClientInterceptor func(ctx context.Context, desc *stream.StreamDesc, method string, streamer Streamer) (stream.ClientStream, error)

// UnaryServerInfo consists of various information about a unary RPC on
// server side. All per-rpc information may be mutated by the interceptor.
type UnaryServerInfo struct {
	// Server is the service implementation the user provides. This is read-only.
	Server interface{}
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
}

// UnaryHandler defines the handler invoked by UnaryServerInterceptor to complete the normal
// execution of a unary RPC.
//
// If a UnaryHandler returns an reason, it should either be produced by the
// errors package, or be one of the context errors. Otherwise, gRPC will use
// codes.Unknown as the errors code and err.Error() as the errors message of the
// RPC.
type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// UnaryServerInterceptor provides a hook to intercept the execution of a unary RPC on the server. info
// contains all the information of this RPC the interceptor can operate on. And handler is the wrapper
// of the service method implementation. It is the responsibility of the interceptor to invoke handler
// to complete the RPC.
type UnaryServerInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)

// StreamServerInfo consists of various information about a streaming RPC on
// server side. All per-rpc information may be mutated by the interceptor.
type StreamServerInfo struct {
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
	// IsClientStream indicates whether the RPC is a client streaming RPC.
	IsClientStream bool
	// IsServerStream indicates whether the RPC is a server streaming RPC.
	IsServerStream bool
}

// StreamServerInterceptor provides a hook to intercept the execution of a streaming RPC on the server.
// info contains all the information of this RPC the interceptor can operate on. And handler is the
// service method implementation. It is the responsibility of the interceptor to invoke handler to
// complete the RPC.
type StreamServerInterceptor func(srv interface{}, ss stream.ServerStream, info *StreamServerInfo, handler stream.StreamHandler) error

type (
	UnaryClientIntBuilder  func(string) UnaryClientInterceptor
	StreamClientIntBuilder func(string) StreamClientInterceptor
	UnaryServerIntBuilder  func() UnaryServerInterceptor
	StreamServerIntBuilder func() StreamServerInterceptor
)

var (
	mu                     sync.RWMutex
	unaryClientIntBuilder  = map[string]UnaryClientIntBuilder{}
	unaryServerIntBuilder  = map[string]UnaryServerIntBuilder{}
	streamClientIntBuilder = map[string]StreamClientIntBuilder{}
	streamServerIntBuilder = map[string]StreamServerIntBuilder{}
)

func RegisterUnaryClientIntBuilder(name string, f UnaryClientIntBuilder) {
	mu.Lock()
	defer mu.Unlock()
	unaryClientIntBuilder[name] = f
}

func RegisterUnaryServerIntBuilder(name string, f UnaryServerIntBuilder) {
	mu.Lock()
	defer mu.Unlock()
	unaryServerIntBuilder[name] = f
}

func RegisterStreamClientIntBuilder(name string, f StreamClientIntBuilder) {
	mu.Lock()
	defer mu.Unlock()
	streamClientIntBuilder[name] = f
}

func RegisterStreamServerIntBuilder(name string, f StreamServerIntBuilder) {
	mu.Lock()
	defer mu.Unlock()
	streamServerIntBuilder[name] = f
}

func getUnaryClientIntBuilder(name string) UnaryClientIntBuilder {
	mu.RLock()
	defer mu.RUnlock()
	f, _ := unaryClientIntBuilder[name]
	return f
}

func getUnaryServerIntBuilder(name string) UnaryServerIntBuilder {
	mu.RLock()
	defer mu.RUnlock()
	f, _ := unaryServerIntBuilder[name]
	return f
}

func getStreamClientIntBuilder(name string) StreamClientIntBuilder {
	mu.RLock()
	defer mu.RUnlock()
	f, _ := streamClientIntBuilder[name]
	return f
}

func getStreamServerIntBuilder(name string) StreamServerIntBuilder {
	mu.RLock()
	defer mu.RUnlock()
	f, _ := streamServerIntBuilder[name]
	return f
}

// ChainUnaryClientInterceptors chains all unary client interceptors into one.
func ChainUnaryClientInterceptors(serviceName string, names []string) UnaryClientInterceptor {
	if len(names) == 0 {
		return func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error {
			return invoker(ctx, method, req, reply)
		}
	}
	if len(names) == 1 {
		return getUnaryClientIntBuilder(names[0])(serviceName)
	}
	interceptors := make([]UnaryClientInterceptor, len(names))
	for _, item := range names {
		interceptors = append(interceptors, getUnaryClientIntBuilder(item)(serviceName))
	}
	return func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error {
		return interceptors[0](ctx, method, req, reply, getChainUnaryInvoker(interceptors, 0, invoker))
	}
}

// getChainUnaryInvoker recursively generate the chained unary invoker.
func getChainUnaryInvoker(interceptors []UnaryClientInterceptor, curr int, finalInvoker UnaryInvoker) UnaryInvoker {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, method string, req, reply interface{}) error {
		return interceptors[curr+1](ctx, method, req, reply, getChainUnaryInvoker(interceptors, curr+1, finalInvoker))
	}
}

// ChainStreamClientInterceptors chains all stream client interceptors into one.
func ChainStreamClientInterceptors(serviceName string, names []string) StreamClientInterceptor {
	if len(names) == 0 {
		return func(ctx context.Context, desc *stream.StreamDesc, method string, streamer Streamer) (stream.ClientStream, error) {
			return streamer(ctx, desc, method)
		}
	}
	if len(names) == 1 {
		return getStreamClientIntBuilder(names[0])(serviceName)
	}
	interceptors := make([]StreamClientInterceptor, len(names))
	for _, item := range names {
		interceptors = append(interceptors, getStreamClientIntBuilder(item)(serviceName))
	}
	return func(ctx context.Context, desc *stream.StreamDesc, method string, streamer Streamer) (stream.ClientStream, error) {
		return interceptors[0](ctx, desc, method, getChainStreamer(interceptors, 0, streamer))
	}
}

// getChainStreamer recursively generate the chained client stream constructor.
func getChainStreamer(interceptors []StreamClientInterceptor, curr int, finalStreamer Streamer) Streamer {
	if curr == len(interceptors)-1 {
		return finalStreamer
	}
	return func(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
		return interceptors[curr+1](ctx, desc, method, getChainStreamer(interceptors, curr+1, finalStreamer))
	}
}

// ChainUnaryServerInterceptors chains all unary server interceptors into one.
func ChainUnaryServerInterceptors(names []string) UnaryServerInterceptor {
	interceptors := make([]UnaryServerInterceptor, len(names))
	for _, item := range names {
		builder := getUnaryServerIntBuilder(item)
		if builder == nil {
			logger.WarnFiled("not found unary server interceptor", logger.String("name", names[0]))
			continue
		}
		interceptors = append(interceptors, builder())
	}
	if len(interceptors) == 0 {
		return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	if len(interceptors) == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
		return interceptors[0](ctx, req, info, getChainUnaryHandler(interceptors, 0, info, handler))
	}
}

func getChainUnaryHandler(interceptors []UnaryServerInterceptor, curr int, info *UnaryServerInfo, finalHandler UnaryHandler) UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}

// ChainStreamServerInterceptors chains all stream server interceptors into one.
func ChainStreamServerInterceptors(names []string) StreamServerInterceptor {
	if len(names) == 0 {
		return func(srv interface{}, ss stream.ServerStream, info *StreamServerInfo, handler stream.StreamHandler) error {
			return handler(srv, ss)
		}
	}
	if len(names) == 1 {
		return getStreamServerIntBuilder(names[0])()
	}

	interceptors := make([]StreamServerInterceptor, len(names))
	for _, item := range names {
		interceptors = append(interceptors, getStreamServerIntBuilder(item)())
	}
	return func(srv interface{}, ss stream.ServerStream, info *StreamServerInfo, handler stream.StreamHandler) error {
		return interceptors[0](srv, ss, info, getChainStreamHandler(interceptors, 0, info, handler))
	}
}

func getChainStreamHandler(interceptors []StreamServerInterceptor, curr int, info *StreamServerInfo, finalHandler stream.StreamHandler) stream.StreamHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(srv interface{}, stream stream.ServerStream) error {
		return interceptors[curr+1](srv, stream, info, getChainStreamHandler(interceptors, curr+1, info, finalHandler))
	}
}
