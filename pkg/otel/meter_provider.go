package otel

import (
	"go.opentelemetry.io/otel/metric"
)

type MeterProviderBuilder func(name string) metric.MeterProvider

var meterBuilders = make(map[string]MeterProviderBuilder)

func RegisterMeterProviderBuilder(name string, constructor MeterProviderBuilder) {
	meterBuilders[name] = constructor
}

func GetMeterProviderBuilder(name string) MeterProviderBuilder {
	constructor, _ := meterBuilders[name]
	return constructor
}
