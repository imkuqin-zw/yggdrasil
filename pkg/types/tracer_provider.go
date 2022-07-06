package types

import "go.opentelemetry.io/otel/trace"

type TracerProviderConstructor func(name string) trace.TracerProvider
