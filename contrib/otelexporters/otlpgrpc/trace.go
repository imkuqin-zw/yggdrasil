package otlpgrpc

import (
	"context"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.1"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	otel.RegisterTracerProviderBuilder("otlpgrpc", tracerProviderBuild)
}

func tracerProviderBuild(_ string) trace.TracerProvider {
	initGrpcConn()
	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName(pkg.Name()),
			semconv.ServiceNamespace(pkg.Namespace()),
		),
	)
	if err != nil {
		logger.FatalField("failed to create resource", logger.Err(err))
		return nil
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(grpcConn))
	if err != nil {
		logger.FatalField("failed to create trace exporter", logger.Err(err))
		return nil
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	defers.Register(func() error {
		cctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		return tracerProvider.Shutdown(cctx)
	})
	return tracerProvider
}
