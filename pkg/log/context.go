package log

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func encodeContext(ctx context.Context, enc ObjectEncoder) error {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		enc.AddString("traceID", spanCtx.TraceID().String())
		enc.AddString("spanID", spanCtx.SpanID().String())
	}
	return nil
}
