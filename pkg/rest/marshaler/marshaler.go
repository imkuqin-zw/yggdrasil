package marshaler

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

var marshalerBuilder = map[string]MarshallerBuilder{}

func RegisterMarshallerBuilder(name string, builder MarshallerBuilder) {
	marshalerBuilder[name] = builder
}

var marshaler = map[string]Marshaler{}

func getMarshaller(name string) Marshaler {
	if marshaler, ok := marshaler[name]; ok {
		return marshaler
	}
	f, ok := marshalerBuilder[name]
	if !ok {
		logger.FatalField("rest marshaler  not found", logger.String("name", name))
	}
	return f()
}

type (
	inbound  = struct{}
	outbound = struct{}
)

func InboundFromContext(ctx context.Context) Marshaler {
	m, ok := ctx.Value(inbound{}).(Marshaler)
	if !ok {
		return defaultMarshaler
	}
	return m
}

func WithInboundContext(ctx context.Context, m Marshaler) context.Context {
	return context.WithValue(ctx, inbound{}, m)
}

func OutboundFromContext(ctx context.Context) Marshaler {
	m, ok := ctx.Value(outbound{}).(Marshaler)
	if !ok {
		return defaultMarshaler
	}
	return m
}

func WithOutboundContext(ctx context.Context, m Marshaler) context.Context {
	return context.WithValue(ctx, outbound{}, m)
}
