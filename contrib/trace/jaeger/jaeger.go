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

package jaeger

import (
	"context"
	"log"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	trace2 "github.com/imkuqin-zw/yggdrasil/pkg/trace"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	trace2.RegisterConstructor("jaeger", Build)
}

type Config struct {
	name      string
	namespace string
	Endpoint  string
	Sampler   float64
}

func (config *Config) Build() trace.TracerProvider {
	if config.name == "" {
		log.Fatal("jaeger name can not be empty")
		return nil
	}
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.Endpoint)))
	if err != nil {
		log.Fatalf("fault to new jaeger collector, err: %+v", err)
		return nil
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(config.Sampler))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(config.name),
			semconv.ServiceNamespaceKey.String(config.namespace),
		)),
	)
	defers.Register(func() error {
		return tp.ForceFlush(context.Background())
	})

	return tp
}

func Build(name string) trace.TracerProvider {
	cfg := &Config{}
	if err := config.Scan("jaeger", cfg); err != nil {
		log.Fatalf("fault to load jaeger config, err: %s", err.Error())
		return nil
	}
	cfg.name = name
	cfg.namespace = config.GetString("yggdrasil.application.namespace")
	return cfg.Build()
}
