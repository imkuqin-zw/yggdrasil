package otel

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/stats"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type clientHandler struct {
	handler
}

func newCliHandler() stats.Handler {
	h := &clientHandler{
		handler: newHandler(false),
	}
	return h
}

// TagRPC can attach some information to the given context.
func (h *clientHandler) TagRPC(ctx context.Context, info stats.RPCTagInfo) context.Context {
	spanName, attrs := parseFullMethod(info.GetFullMethod())
	attrs = append(attrs, semconv.RPCSystemKey.String("yggdrasil"))
	ctx, _ = h.tracer.Start(
		ctx,
		spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)

	gctx := rpcContext{
		metricAttrs: attrs,
	}

	return inject(context.WithValue(ctx, rpcContextKey{}, &gctx), otel.GetTextMapPropagator())
}

func (h *clientHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	h.handleRPC(ctx, rs, false)
}

func (h *clientHandler) TagChannel(ctx context.Context, info stats.ChanTagInfo) context.Context {
	return ctx
}

func (h *clientHandler) HandleChannel(context.Context, stats.ChanStats) {
	// no-op
}
