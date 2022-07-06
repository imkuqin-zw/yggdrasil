package trace

import (
	"context"

	grpc2 "github.com/imkuqin-zw/yggdrasil/pkg/client/grpc"
	xtrace "github.com/imkuqin-zw/yggdrasil/pkg/trace"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	grpc2.RegisterUnaryInterceptor("trace", func(string) grpc.UnaryClientInterceptor { return unaryClientInterceptor })
	grpc2.RegisterStreamInterceptor("trace", func(string) grpc.StreamClientInterceptor { return streamClientInterceptor })
}

func unaryClientInterceptor(
	ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) (err error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCMethodKey.String(method),
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	attrs = append(attrs, semconv.RPCMethodKey.String(method))
	ctx, span := tracer.Start(ctx, method, xtrace.MetadataReaderWriter{MD: md}, trace.WithAttributes(attrs...))
	defer tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return invoker(ctx, method, req, reply, cc, opts...)
}

func streamClientInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (cs grpc.ClientStream, err error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		semconv.RPCMethodKey.String(method),
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	attrs = append(attrs, semconv.RPCMethodKey.String(method))
	ctx, span := tracer.Start(ctx, method, xtrace.MetadataReaderWriter{MD: md}, trace.WithAttributes(attrs...))
	ctx = metadata.NewOutgoingContext(ctx, md)
	cs, err = streamer(ctx, desc, cc, method, opts...)
	stream := wrapClientStream(ctx, cs, desc)
	go func() {
		err := <-stream.finished
		tracer.End(ctx, span, semconv.RPCGRPCStatusCodeKey, err)
	}()
	return stream, nil
}
