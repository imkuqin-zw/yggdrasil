package otel

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/stats"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/genproto/googleapis/rpc/code"
)

type rpcContextKey struct{}

type rpcContext struct {
	messagesReceived int64
	messagesSent     int64
	metricAttrs      []attribute.KeyValue
}

var (
	messageTransportSizeKey = attribute.Key("message.transport_size")
	messageDataSizeKey      = attribute.Key("message.data_size")
	peerEndpointKey         = attribute.Key("peer_endpoint")
	protocolKey             = attribute.Key("protocol")
	codeKey                 = attribute.Key("rpc.status_code")
)

type handler struct {
	cfg                *Config
	tracer             trace.Tracer
	rpcDuration        metric.Float64Histogram
	rpcRequestSize     metric.Int64Histogram
	rpcResponseSize    metric.Int64Histogram
	rpcRequestsPerRPC  metric.Int64Histogram
	rpcResponsesPerRPC metric.Int64Histogram

	handleRPC func(context.Context, stats.RPCStats, bool)
}

func newHandler(isSvr bool) handler {
	cfg := getCfg()
	tracer := otel.Tracer("github.com/imkuqin-zw/yggdrasil", trace.WithInstrumentationVersion("semver:"+pkg.FrameworkVersion))
	h := handler{
		cfg:    cfg,
		tracer: tracer,
	}
	meter := otel.Meter("github.com/imkuqin-zw/yggdrasil",
		metric.WithInstrumentationVersion("semver:"+pkg.FrameworkVersion),
		metric.WithSchemaURL(semconv.SchemaURL),
	)

	if cfg.EnableMetrics {
		var role string
		if isSvr {
			role = "server"
		} else {
			role = "client"
		}
		var err error
		h.rpcDuration, err = meter.Float64Histogram("rpc."+role+".duration",
			metric.WithDescription("Measures the duration of inbound RPC."),
			metric.WithUnit("ms"))
		if err != nil {
			otel.Handle(err)
			if h.rpcDuration == nil {
				h.rpcDuration = noop.Float64Histogram{}
			}
		}

		h.rpcRequestSize, err = meter.Int64Histogram("rpc."+role+".request.size",
			metric.WithDescription("Measures size of RPC request messages (uncompressed)."),
			metric.WithUnit("By"))
		if err != nil {
			otel.Handle(err)
			if h.rpcRequestSize == nil {
				h.rpcRequestSize = noop.Int64Histogram{}
			}
		}

		h.rpcResponseSize, err = meter.Int64Histogram("rpc."+role+".response.size",
			metric.WithDescription("Measures size of RPC response messages (uncompressed)."),
			metric.WithUnit("By"))
		if err != nil {
			otel.Handle(err)
			if h.rpcResponseSize == nil {
				h.rpcResponseSize = noop.Int64Histogram{}
			}
		}

		h.rpcRequestsPerRPC, err = meter.Int64Histogram("rpc."+role+".requests_per_rpc",
			metric.WithDescription("Measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs."),
			metric.WithUnit("{count}"))
		if err != nil {
			otel.Handle(err)
			if h.rpcRequestsPerRPC == nil {
				h.rpcRequestsPerRPC = noop.Int64Histogram{}
			}
		}

		h.rpcResponsesPerRPC, err = meter.Int64Histogram("rpc."+role+".responses_per_rpc",
			metric.WithDescription("Measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs."),
			metric.WithUnit("{count}"))
		if err != nil {
			otel.Handle(err)
			if h.rpcResponsesPerRPC == nil {
				h.rpcResponsesPerRPC = noop.Int64Histogram{}
			}
		}
		h.handleRPC = h.handleWithMetrics
	} else {
		h.handleRPC = h.handleWithOutMetrics
	}

	return h
}

func (h *handler) handleWithMetrics(ctx context.Context, rs stats.RPCStats, isServer bool) {
	span := trace.SpanFromContext(ctx)
	rctx, _ := ctx.Value(rpcContextKey{}).(*rpcContext)
	var metricAttrs []attribute.KeyValue
	if rctx != nil {
		metricAttrs = make([]attribute.KeyValue, 0, len(rctx.metricAttrs)+1)
		metricAttrs = append(metricAttrs, rctx.metricAttrs...)
	}
	var messageId int64
	switch rs := rs.(type) {
	case stats.RPCBegin:
	case stats.RPCInPayload:
		if rctx != nil {
			messageId = atomic.AddInt64(&rctx.messagesReceived, 1)
			h.rpcRequestSize.Record(ctx, int64(rs.GetTransportSize()), metric.WithAttributes(metricAttrs...))
		}

		if h.cfg.ReceivedEvent {
			span.AddEvent("message",
				trace.WithAttributes(
					semconv.MessageTypeReceived,
					semconv.MessageIDKey.Int64(messageId),
					messageTransportSizeKey.Int(rs.GetTransportSize()),
					messageDataSizeKey.Int(len(rs.GetData())),
				),
			)
		}
	case stats.RPCOutPayload:
		if rctx != nil {
			messageId = atomic.AddInt64(&rctx.messagesSent, 1)
			h.rpcResponseSize.Record(ctx, int64(rs.GetTransportSize()), metric.WithAttributes(metricAttrs...))
		}

		if h.cfg.SentEvent {
			span.AddEvent("message",
				trace.WithAttributes(
					semconv.MessageTypeSent,
					semconv.MessageIDKey.Int64(messageId),
					protocolKey.String(rs.GetProtocol()),
					messageTransportSizeKey.Int(rs.GetTransportSize()),
					messageDataSizeKey.Int(len(rs.GetData())),
				),
			)
		}
	case stats.RPCOutTrailer:
	case stats.RPCOutHeader:
		span.SetAttributes(protocolKey.String(rs.GetProtocol()))
		span.SetAttributes(peerEndpointKey.String(rs.GetRemoteEndpoint()))
	case stats.RPCEnd:
		var rpcStatusAttr attribute.KeyValue

		if rs.Error() != nil {
			s := status.FromError(rs.Error())
			if isServer {
				statusCode, msg := serverStatus(s)
				span.SetStatus(statusCode, msg)
			} else {
				span.SetStatus(codes.Error, s.Message())
			}
			rpcStatusAttr = codeKey.Int(int(s.Code()))
		} else {
			rpcStatusAttr = codeKey.Int(int(code.Code_OK))
		}
		span.SetAttributes(rpcStatusAttr)
		span.End()

		metricAttrs = append(metricAttrs, rpcStatusAttr)

		// Use floating point division here for higher precision (instead of Millisecond method).
		elapsedTime := float64(rs.GetEndTime().Sub(rs.GetBeginTime())) / float64(time.Millisecond)

		h.rpcDuration.Record(ctx, elapsedTime, metric.WithAttributes(metricAttrs...))
		if rctx != nil {
			h.rpcRequestsPerRPC.Record(ctx, atomic.LoadInt64(&rctx.messagesReceived), metric.WithAttributes(metricAttrs...))
			h.rpcResponsesPerRPC.Record(ctx, atomic.LoadInt64(&rctx.messagesSent), metric.WithAttributes(metricAttrs...))
		}
	default:
		return
	}
}

func (h *handler) handleWithOutMetrics(ctx context.Context, rs stats.RPCStats, isServer bool) {
	span := trace.SpanFromContext(ctx)
	rctx, _ := ctx.Value(rpcContextKey{}).(*rpcContext)
	var messageId int64
	switch rs := rs.(type) {
	case stats.RPCBegin:
	case stats.RPCInPayload:
		if rctx != nil {
			messageId = atomic.AddInt64(&rctx.messagesReceived, 1)
		}

		if h.cfg.ReceivedEvent {
			span.AddEvent("message",
				trace.WithAttributes(
					semconv.MessageTypeReceived,
					semconv.MessageIDKey.Int64(messageId),
					messageTransportSizeKey.Int(rs.GetTransportSize()),
					messageDataSizeKey.Int(len(rs.GetData())),
				),
			)
		}
	case stats.RPCOutPayload:
		if rctx != nil {
			messageId = atomic.AddInt64(&rctx.messagesSent, 1)
		}

		if h.cfg.SentEvent {
			span.AddEvent("message",
				trace.WithAttributes(
					semconv.MessageTypeSent,
					semconv.MessageIDKey.Int64(messageId),
					protocolKey.String(rs.GetProtocol()),
					messageTransportSizeKey.Int(rs.GetTransportSize()),
					messageDataSizeKey.Int(len(rs.GetData())),
				),
			)
		}
	case stats.RPCOutTrailer:
	case stats.RPCOutHeader:
		span.SetAttributes(protocolKey.String(rs.GetProtocol()))
		span.SetAttributes(peerEndpointKey.String(rs.GetRemoteEndpoint()))
	case stats.RPCEnd:
		var rpcStatusAttr attribute.KeyValue

		if rs.Error() != nil {
			s := status.FromError(rs.Error())
			if isServer {
				statusCode, msg := serverStatus(s)
				span.SetStatus(statusCode, msg)
			} else {
				span.SetStatus(codes.Error, s.Message())
			}
			rpcStatusAttr = codeKey.Int(int(s.Code()))
		} else {
			rpcStatusAttr = codeKey.Int(int(code.Code_OK))
		}
		span.SetAttributes(rpcStatusAttr)
		span.End()
	default:
		return
	}
}

func serverStatus(st *status.Status) (codes.Code, string) {
	switch code.Code(st.Code()) {
	case code.Code_UNKNOWN,
		code.Code_INTERNAL,
		code.Code_DEADLINE_EXCEEDED,
		code.Code_UNIMPLEMENTED,
		code.Code_UNAVAILABLE,
		code.Code_DATA_LOSS:
		return codes.Error, st.Message()
	default:
		return codes.Unset, ""
	}
}
