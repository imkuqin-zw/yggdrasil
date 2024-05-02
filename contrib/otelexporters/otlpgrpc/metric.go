package otlpgrpc

import (
	"context"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.1"
)

func init() {
	otel.RegisterMeterProviderBuilder("otlpgrpc", buildMeterProvider)
}

func buildMeterProvider(_ string) metric.MeterProvider {
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
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(grpcConn))
	if err != nil {
		logger.FatalField("failed to create trace exporter", logger.Err(err))
		return nil
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	defers.Register(func() error {
		cctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		return meterProvider.Shutdown(cctx)
	})
	return meterProvider
}
