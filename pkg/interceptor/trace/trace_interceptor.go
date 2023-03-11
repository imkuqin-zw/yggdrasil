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

package trace

import (
	"context"
	"strings"

	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/peer"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	xtrace "github.com/imkuqin-zw/yggdrasil/pkg/tracer"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

var name = "trace"

func init() {
	interceptor.RegisterUnaryClientIntBuilder(name, func(s string) interceptor.UnaryClientInterceptor {
		return unaryClientInterceptor
	})
	interceptor.RegisterStreamClientIntBuilder(name, func(s string) interceptor.StreamClientInterceptor {
		return streamClientInterceptor
	})
	interceptor.RegisterUnaryServerIntBuilder(name, func() interceptor.UnaryServerInterceptor {
		return unaryServerInterceptor
	})
	interceptor.RegisterStreamServerIntBuilder(name, func() interceptor.StreamServerInterceptor {
		return streamServerInterceptor
	})
}

func unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, invoker interceptor.UnaryInvoker) (err error) {
	md, _ := metadata.FromOutContext(ctx)
	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCMethodKey.String(method),
	}
	ctx = metadata.WithOutContext(ctx, md)
	attrs = append(attrs, semconv.RPCMethodKey.String(method))
	ctx, span := tracer.Start(ctx, method, xtrace.MetadataReaderWriter{MD: md}, trace.WithAttributes(attrs...))
	defer tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	ctx = metadata.WithOutContext(ctx, md)
	return invoker(ctx, method, req, reply)
}

func streamClientInterceptor(ctx context.Context, desc *stream.StreamDesc, method string, streamer interceptor.Streamer) (cs stream.ClientStream, err error) {
	md, _ := metadata.FromOutContext(ctx)
	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCMethodKey.String(method),
	}
	ctx = metadata.WithOutContext(ctx, md)
	attrs = append(attrs, semconv.RPCMethodKey.String(method))
	ctx, span := tracer.Start(ctx, method, xtrace.MetadataReaderWriter{MD: md}, trace.WithAttributes(attrs...))
	ctx = metadata.WithOutContext(ctx, md)
	cs, err = streamer(ctx, desc, method)
	stream := wrapClientStream(ctx, cs, desc)
	go func() {
		err := <-stream.finished
		tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	}()
	return stream, nil
}

func unaryServerInterceptor(ctx context.Context, req interface{}, info *interceptor.UnaryServerInfo, handler interceptor.UnaryHandler) (reply interface{}, err error) {
	md, _ := metadata.FromInContext(ctx)
	operation := strings.TrimLeft(info.FullMethod, "/")
	parts := strings.SplitN(operation, "/", 2)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCServiceKey.String(parts[0]),
		semconv.RPCMethodKey.String(parts[1]),
	}
	p, ok := peer.PeerFromContext(ctx)
	if ok {
		attrs = append(attrs, xtrace.PeerAttr(p.Addr.String())...)
	}
	tracer := xtrace.NewTracer(trace.SpanKindServer)
	ctx, span := tracer.Start(ctx, operation, xtrace.MetadataReaderWriter{MD: md})
	defer tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	return handler(ctx, req)
}

func streamServerInterceptor(srv interface{}, ss stream.ServerStream, info *interceptor.StreamServerInfo, handler stream.StreamHandler) (err error) {
	md, _ := metadata.FromInContext(ss.Context())
	operation := strings.TrimLeft(info.FullMethod, "/")
	parts := strings.SplitN(operation, "/", 2)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCServiceKey.String(parts[0]),
		semconv.RPCMethodKey.String(parts[1]),
	}
	p, ok := peer.PeerFromContext(ss.Context())
	if ok {
		attrs = append(attrs, xtrace.PeerAttr(p.Addr.String())...)
	}
	tracer := xtrace.NewTracer(trace.SpanKindServer)
	ctx, span := tracer.Start(ss.Context(), operation, xtrace.MetadataReaderWriter{MD: md}, trace.WithAttributes(attrs...))
	defer tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	return handler(srv, &serverStream{
		ServerStream: ss,
		ctx:          ctx,
	})
}
