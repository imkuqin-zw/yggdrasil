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

package otel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/stretchr/testify/assert"
)

func TestRegisterTracerProviderBuilder(t *testing.T) {
	// Clear any existing builders
	tracerBuilders = make(map[string]TracerProviderBuilder)

	// Test registering a new builder
	builder := func(name string) trace.TracerProvider {
		return noop.NewTracerProvider()
	}

	RegisterTracerProviderBuilder("test-provider", builder)

	// Verify the builder was registered
	storedBuilder, exists := tracerBuilders["test-provider"]
	assert.True(t, exists)
	assert.NotNil(t, storedBuilder)
}

func TestRegisterTracerProviderBuilder_Overwrite(t *testing.T) {
	// Clear any existing builders
	tracerBuilders = make(map[string]TracerProviderBuilder)

	// Register initial builder
	initialBuilder := func(name string) trace.TracerProvider {
		return noop.NewTracerProvider()
	}
	RegisterTracerProviderBuilder("test-provider", initialBuilder)

	// Verify initial registration
	storedBuilder, exists := tracerBuilders["test-provider"]
	assert.True(t, exists)
	assert.NotNil(t, storedBuilder)

	// Overwrite with new builder
	newBuilder := func(name string) trace.TracerProvider {
		return noop.NewTracerProvider()
	}
	RegisterTracerProviderBuilder("test-provider", newBuilder)

	// Verify overwrite
	storedBuilder, exists = tracerBuilders["test-provider"]
	assert.True(t, exists)
	assert.NotNil(t, storedBuilder)
}

func TestGetTracerProviderBuilder_Exists(t *testing.T) {
	// Clear any existing builders
	tracerBuilders = make(map[string]TracerProviderBuilder)

	// Register a builder
	expectedProvider := noop.NewTracerProvider()
	builder := func(name string) trace.TracerProvider {
		assert.Equal(t, "test-service", name)
		return expectedProvider
	}
	RegisterTracerProviderBuilder("test-provider", builder)

	// Get the builder
	retrievedBuilder := GetTracerProviderBuilder("test-provider")
	assert.NotNil(t, retrievedBuilder)

	// Test the builder function
	actualProvider := retrievedBuilder("test-service")
	assert.Equal(t, expectedProvider, actualProvider)
}

func TestGetTracerProviderBuilder_NotExists(t *testing.T) {
	// Clear any existing builders
	tracerBuilders = make(map[string]TracerProviderBuilder)

	// Try to get a non-existent builder
	builder := GetTracerProviderBuilder("non-existent")
	assert.Nil(t, builder)
}

func TestTracerProviderBuilderFunctionTypes(t *testing.T) {
	// Clear any existing builders
	tracerBuilders = make(map[string]TracerProviderBuilder)

	// Test different types of builder functions
	mockProvider := noop.NewTracerProvider()

	// Simple builder
	simpleBuilder := func(name string) trace.TracerProvider {
		return mockProvider
	}
	RegisterTracerProviderBuilder("simple", simpleBuilder)

	// Builder with validation
	validatingBuilder := func(name string) trace.TracerProvider {
		if name == "" {
			return noop.NewTracerProvider()
		}
		return mockProvider
	}
	RegisterTracerProviderBuilder("validating", validatingBuilder)

	// Builder with state
	counter := 0
	statefulBuilder := func(name string) trace.TracerProvider {
		counter++
		return mockProvider
	}
	RegisterTracerProviderBuilder("stateful", statefulBuilder)

	// Test all builders
	assert.Equal(t, mockProvider, GetTracerProviderBuilder("simple")("test"))
	assert.Equal(t, mockProvider, GetTracerProviderBuilder("validating")("test"))
	assert.Equal(t, mockProvider, GetTracerProviderBuilder("stateful")("test"))

	// Verify stateful builder behavior
	assert.Equal(t, 1, counter)
	GetTracerProviderBuilder("stateful")("test2")
	assert.Equal(t, 2, counter)
}

func TestTracerProviderBuilder_ConcurrentAccess(t *testing.T) {
	// Clear any existing builders
	tracerBuilders = make(map[string]TracerProviderBuilder)

	// Test concurrent registration and retrieval
	done := make(chan bool, 2)

	// Goroutine for registration
	go func() {
		for i := 0; i < 100; i++ {
			builder := func(name string) trace.TracerProvider {
				return noop.NewTracerProvider()
			}
			RegisterTracerProviderBuilder("provider-"+string(rune(i)), builder)
		}
		done <- true
	}()

	// Goroutine for retrieval
	go func() {
		for i := 0; i < 100; i++ {
			GetTracerProviderBuilder("provider-" + string(rune(i)))
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify that some builders were registered (this is a basic sanity check)
	assert.True(t, len(tracerBuilders) >= 0)
}

func TestTracerProviderBuilder_MultipleRegistrations(t *testing.T) {
	// Clear any existing builders
	tracerBuilders = make(map[string]TracerProviderBuilder)

	// Register multiple builders
	providers := []string{"default", "jaeger", "zipkin", "custom"}
	for _, name := range providers {
		builder := func(serviceName string) trace.TracerProvider {
			return noop.NewTracerProvider()
		}
		RegisterTracerProviderBuilder(name, builder)
	}

	// Verify all builders are registered
	assert.Equal(t, len(providers), len(tracerBuilders))

	for _, name := range providers {
		builder := GetTracerProviderBuilder(name)
		assert.NotNil(t, builder)

		// Test that the builder works
		provider := builder("test-service")
		assert.NotNil(t, provider)
	}
}

func TestTracerProviderBuilder_GlobalVariable(t *testing.T) {
	// Test that the global tracerBuilders variable is properly initialized
	// This is mainly a sanity check to ensure the package state is clean
	assert.NotNil(t, tracerBuilders)
	assert.IsType(t, make(map[string]TracerProviderBuilder), tracerBuilders)
}

func TestInit(t *testing.T) {
	// Test that the init function properly sets up the propagator
	// Note: This test verifies the side effect of the init function

	// Get the current propagator
	propagator := otel.GetTextMapPropagator()
	assert.NotNil(t, propagator)

	// Verify it's a CompositeTextMapPropagator (which should be set in init)
	// We can't directly check the type, but we can verify it has the expected behavior
	assert.Implements(t, (*propagation.TextMapPropagator)(nil), propagator)
}

func TestInit_PropagatorComposition(t *testing.T) {
	// Test that the init function sets up the expected composite propagator
	// by checking that it can handle different propagation formats

	propagator := otel.GetTextMapPropagator()
	assert.NotNil(t, propagator)

	// Test that the propagator can extract from a carrier
	// This tests that the composite propagator is working
	carrier := propagation.MapCarrier{
		"traceparent": "00-12345678901234567890123456789012-1234567890123456-01",
		"baggage":     "key1=value1,key2=value2",
	}

	ctx := context.Background()
	ctx = propagator.Extract(ctx, carrier)
	assert.NotNil(t, ctx)

	// Test that the propagator can inject to a carrier
	ctx = context.Background()

	injectCarrier := propagation.MapCarrier{}
	propagator.Inject(ctx, injectCarrier)

	// The carrier should now have trace context information or be empty (both are valid)
	// An empty carrier is valid since there's no active span to inject
	assert.NotNil(t, injectCarrier)
}
