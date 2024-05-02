package otel

import (
	"context"
	"strings"

	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	xtrace "github.com/imkuqin-zw/yggdrasil/pkg/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.1"
)

func parseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
	if !strings.HasPrefix(fullMethod, "/") {
		// Invalid format, does not follow `/package.service/method`.
		return fullMethod, nil
	}
	name := fullMethod[1:]
	pos := strings.LastIndex(name, "/")
	if pos < 0 {
		// Invalid format, does not follow `/package.service/method`.
		return name, nil
	}
	service, method := name[:pos], name[pos+1:]

	var attrs []attribute.KeyValue
	if service != "" {
		attrs = append(attrs, semconv.RPCService(service))
	}
	if method != "" {
		attrs = append(attrs, semconv.RPCMethod(method))
	}
	return name, attrs
}

func inject(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, _ := metadata.FromOutContext(ctx)
	propagators.Inject(ctx, xtrace.NewMetadataReaderWriter(&md))
	return metadata.WithOutContext(ctx, md)
}

func extract(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, _ := metadata.FromInContext(ctx)
	return propagators.Extract(ctx, xtrace.NewMetadataReaderWriter(&md))
}
