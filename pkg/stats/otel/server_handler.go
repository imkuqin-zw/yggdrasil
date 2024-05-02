package otel

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/stats"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type serverHandler struct {
	handler
}

func newSvrHandler() *serverHandler {
	return &serverHandler{
		handler: newHandler(true),
	}
}

func (h *serverHandler) TagChannel(ctx context.Context, info stats.ChanTagInfo) context.Context {
	return ctx
}

func (h *serverHandler) HandleChannel(ctx context.Context, info stats.ChanStats) {
}

// TagRPC can attach some information to the given context.
func (h *serverHandler) TagRPC(ctx context.Context, info stats.RPCTagInfo) context.Context {
	ctx = extract(ctx, otel.GetTextMapPropagator())

	spanName, attrs := parseFullMethod(info.GetFullMethod())
	attrs = append(attrs, semconv.RPCSystemKey.String("yggdrasil"))
	ctx, _ = h.tracer.Start(
		trace.ContextWithRemoteSpanContext(ctx, trace.SpanContextFromContext(ctx)),
		spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...),
	)

	gctx := rpcContext{
		metricAttrs: attrs,
	}
	return context.WithValue(ctx, rpcContextKey{}, &gctx)
}

// HandleRPC processes the RPC stats.
func (h *serverHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	isServer := true
	h.handleRPC(ctx, rs, isServer)
}
