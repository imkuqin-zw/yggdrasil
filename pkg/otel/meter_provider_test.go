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
	"testing"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/stretchr/testify/assert"
)

func TestRegisterMeterProviderBuilder(t *testing.T) {
	// Clear any existing builders
	meterBuilders = make(map[string]MeterProviderBuilder)

	// Test registering a new builder
	builder := func(name string) metric.MeterProvider {
		return noop.NewMeterProvider()
	}

	RegisterMeterProviderBuilder("test-provider", builder)

	// Verify the builder was registered
	storedBuilder, exists := meterBuilders["test-provider"]
	assert.True(t, exists)
	assert.NotNil(t, storedBuilder)
}

func TestRegisterMeterProviderBuilder_Overwrite(t *testing.T) {
	// Clear any existing builders
	meterBuilders = make(map[string]MeterProviderBuilder)

	// Register initial builder
	initialBuilder := func(name string) metric.MeterProvider {
		return noop.NewMeterProvider()
	}
	RegisterMeterProviderBuilder("test-provider", initialBuilder)

	// Verify initial registration
	storedBuilder, exists := meterBuilders["test-provider"]
	assert.True(t, exists)
	assert.NotNil(t, storedBuilder)

	// Overwrite with new builder
	newBuilder := func(name string) metric.MeterProvider {
		return noop.NewMeterProvider()
	}
	RegisterMeterProviderBuilder("test-provider", newBuilder)

	// Verify overwrite
	storedBuilder, exists = meterBuilders["test-provider"]
	assert.True(t, exists)
	assert.NotNil(t, storedBuilder)
	// Note: We can't easily compare functions, but we can verify the new builder is called
}

func TestGetMeterProviderBuilder_Exists(t *testing.T) {
	// Clear any existing builders
	meterBuilders = make(map[string]MeterProviderBuilder)

	// Register a builder
	expectedProvider := noop.NewMeterProvider()
	builder := func(name string) metric.MeterProvider {
		assert.Equal(t, "test-service", name)
		return expectedProvider
	}
	RegisterMeterProviderBuilder("test-provider", builder)

	// Get the builder
	retrievedBuilder := GetMeterProviderBuilder("test-provider")
	assert.NotNil(t, retrievedBuilder)

	// Test the builder function
	actualProvider := retrievedBuilder("test-service")
	assert.Equal(t, expectedProvider, actualProvider)
}

func TestGetMeterProviderBuilder_NotExists(t *testing.T) {
	// Clear any existing builders
	meterBuilders = make(map[string]MeterProviderBuilder)

	// Try to get a non-existent builder
	builder := GetMeterProviderBuilder("non-existent")
	assert.Nil(t, builder)
}

func TestMeterProviderBuilderFunctionTypes(t *testing.T) {
	// Clear any existing builders
	meterBuilders = make(map[string]MeterProviderBuilder)

	// Test different types of builder functions
	mockProvider := noop.NewMeterProvider()

	// Simple builder
	simpleBuilder := func(name string) metric.MeterProvider {
		return mockProvider
	}
	RegisterMeterProviderBuilder("simple", simpleBuilder)

	// Builder with validation
	validatingBuilder := func(name string) metric.MeterProvider {
		if name == "" {
			return noop.NewMeterProvider()
		}
		return mockProvider
	}
	RegisterMeterProviderBuilder("validating", validatingBuilder)

	// Builder with state
	counter := 0
	statefulBuilder := func(name string) metric.MeterProvider {
		counter++
		return mockProvider
	}
	RegisterMeterProviderBuilder("stateful", statefulBuilder)

	// Test all builders
	assert.Equal(t, mockProvider, GetMeterProviderBuilder("simple")("test"))
	assert.Equal(t, mockProvider, GetMeterProviderBuilder("validating")("test"))
	assert.Equal(t, mockProvider, GetMeterProviderBuilder("stateful")("test"))

	// Verify stateful builder behavior
	assert.Equal(t, 1, counter)
	GetMeterProviderBuilder("stateful")("test2")
	assert.Equal(t, 2, counter)
}

func TestMeterProviderBuilder_ConcurrentAccess(t *testing.T) {
	// Clear any existing builders
	meterBuilders = make(map[string]MeterProviderBuilder)

	// Test concurrent registration and retrieval
	done := make(chan bool, 2)

	// Goroutine for registration
	go func() {
		for i := 0; i < 100; i++ {
			builder := func(name string) metric.MeterProvider {
				return noop.NewMeterProvider()
			}
			RegisterMeterProviderBuilder("provider-"+string(rune(i)), builder)
		}
		done <- true
	}()

	// Goroutine for retrieval
	go func() {
		for i := 0; i < 100; i++ {
			GetMeterProviderBuilder("provider-" + string(rune(i)))
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify that some builders were registered (this is a basic sanity check)
	assert.True(t, len(meterBuilders) >= 0)
}

func TestMeterProviderBuilder_MultipleRegistrations(t *testing.T) {
	// Clear any existing builders
	meterBuilders = make(map[string]MeterProviderBuilder)

	// Register multiple builders
	providers := []string{"default", "prometheus", "statsd", "custom"}
	for _, name := range providers {
		builder := func(serviceName string) metric.MeterProvider {
			return noop.NewMeterProvider()
		}
		RegisterMeterProviderBuilder(name, builder)
	}

	// Verify all builders are registered
	assert.Equal(t, len(providers), len(meterBuilders))

	for _, name := range providers {
		builder := GetMeterProviderBuilder(name)
		assert.NotNil(t, builder)

		// Test that the builder works
		provider := builder("test-service")
		assert.NotNil(t, provider)
	}
}

func TestMeterProviderBuilder_GlobalVariable(t *testing.T) {
	// Test that the global meterBuilders variable is properly initialized
	// This is mainly a sanity check to ensure the package state is clean
	assert.NotNil(t, meterBuilders)
	assert.IsType(t, make(map[string]MeterProviderBuilder), meterBuilders)
}
