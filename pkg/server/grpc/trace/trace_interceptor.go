package trace

import (
	"context"
	"strings"

	md2 "github.com/imkuqin-zw/yggdrasil/pkg/md"
	grpc2 "github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
	xtrace "github.com/imkuqin-zw/yggdrasil/pkg/trace"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func init() {
	grpc2.RegisterUnaryInterceptor("trace", func() grpc.UnaryServerInterceptor { return traceUnaryServerInterceptor })
	grpc2.RegisterStreamInterceptor("trace", func() grpc.StreamServerInterceptor { return traceStreamServerInterceptor })
}

type tracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *tracingServerStream) Context() context.Context {
	return ss.ctx
}

func traceUnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (reply interface{}, err error) {
	md, ok := md2.FromInContext(ctx)
	if ok {
		md = md.Copy()
	}
	operation := strings.TrimLeft(info.FullMethod, "/")
	parts := strings.SplitN(operation, "/", 2)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCServiceKey.String(parts[0]),
		semconv.RPCMethodKey.String(parts[1]),
	}
	p, ok := peer.FromContext(ctx)
	if ok {
		attrs = append(attrs, xtrace.PeerAttr(p.Addr.String())...)
	}
	tracer := xtrace.NewTracer(trace.SpanKindServer)
	ctx, span := tracer.Start(ctx, operation, xtrace.MetadataReaderWriter{MD: md})
	defer tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	return handler(ctx, req)
}

func traceStreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	md, ok := md2.FromInContext(ss.Context())
	if ok {
		md = md.Copy()
	}
	operation := strings.TrimLeft(info.FullMethod, "/")
	parts := strings.SplitN(operation, "/", 2)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCServiceKey.String(parts[0]),
		semconv.RPCMethodKey.String(parts[1]),
	}
	p, ok := peer.FromContext(ss.Context())
	if ok {
		attrs = append(attrs, xtrace.PeerAttr(p.Addr.String())...)
	}
	tracer := xtrace.NewTracer(trace.SpanKindServer)
	ctx, span := tracer.Start(ss.Context(), operation, xtrace.MetadataReaderWriter{MD: md}, trace.WithAttributes(attrs...))
	defer tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	return handler(srv, &tracingServerStream{
		ServerStream: ss,
		ctx:          ctx,
	})
}
