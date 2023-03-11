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

package tracer

import (
	"context"
	"net"

	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

type options struct {
	propagator propagation.TextMapPropagator
}

// Option is tracing option.
type Option func(*options)

// Tracer is otel span tracer
type Tracer struct {
	tracer trace.Tracer
	kind   trace.SpanKind
	opts   *options
}

// NewTracer create tracer instance
func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	op := options{
		propagator: propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
	}
	for _, o := range opts {
		o(&op)
	}
	return &Tracer{tracer: otel.Tracer("yggdrasil"), kind: kind, opts: &op}
}

// Start  tracing span
func (t *Tracer) Start(ctx context.Context, operation string, carrier propagation.TextMapCarrier, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if (t.kind == trace.SpanKindServer || t.kind == trace.SpanKindConsumer) && carrier != nil {
		ctx = t.opts.propagator.Extract(ctx, carrier)
	}
	ctx, span := t.tracer.Start(ctx,
		operation,
		append(opts, trace.WithSpanKind(t.kind))...,
	)
	if (t.kind == trace.SpanKindClient || t.kind == trace.SpanKindProducer) && carrier != nil {
		t.opts.propagator.Inject(ctx, carrier)
	}
	return ctx, span
}

// End finish tracing span
func (t *Tracer) End(ctx context.Context, span trace.Span, codeKey attribute.Key, err error) {
	if err != nil {
		span.RecordError(err)
		e := status.FromError(err)
		span.SetAttributes(codeKey.Int64(int64(e.Code())))
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}
	span.End()
}

// PeerAttr returns the peer attributes.
func PeerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil
	}

	if len(host) == 0 {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}
