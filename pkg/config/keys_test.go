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

package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "single element",
			input:    []string{"single"},
			expected: "single",
		},
		{
			name:     "two elements",
			input:    []string{"first", "second"},
			expected: "first.second",
		},
		{
			name:     "multiple elements",
			input:    []string{"a", "b", "c", "d"},
			expected: "a.b.c.d",
		},
		{
			name:     "empty strings",
			input:    []string{"", ""},
			expected: ".",
		},
		{
			name:     "mixed with empty",
			input:    []string{"first", "", "third"},
			expected: "first..third",
		},
		{
			name:     "complex keys",
			input:    []string{"yggdrasil", "client", "{service.name}", "config"},
			expected: "yggdrasil.client.{service.name}.config",
		},
		{
			name:     "no elements",
			input:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Join(tt.input...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenPath(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		delimiter string
		expected  []string
	}{
		{
			name:      "simple key",
			key:       "simple.key",
			delimiter: ".",
			expected:  []string{"simple", "key"},
		},
		{
			name:      "key with placeholders",
			key:       "service.{instance.name}.config",
			delimiter: ".",
			expected:  []string{"service", "instance.name", "config"},
		},
		{
			name:      "key with nested placeholders",
			key:       "a.{b.c}.d.{e.f}.g",
			delimiter: ".",
			expected:  []string{"a", "b.c", "d", "e.f", "g"},
		},
		{
			name:      "key with dashes in placeholders",
			key:       "service.{instance-name}.config",
			delimiter: ".",
			expected:  []string{"service", "instance-name", "config"},
		},
		{
			name:      "key with numbers in placeholders",
			key:       "service.{instance123}.config",
			delimiter: ".",
			expected:  []string{"service", "instance123", "config"},
		},
		{
			name:      "key with dots in placeholders",
			key:       "service.{instance.name.test}.config",
			delimiter: ".",
			expected:  []string{"service", "instance.name.test", "config"},
		},
		{
			name:      "single placeholder",
			key:       "{single.placeholder}",
			delimiter: ".",
			expected:  []string{"single.placeholder"},
		},
		{
			name:      "no placeholders",
			key:       "simple.path.key",
			delimiter: ".",
			expected:  []string{"simple", "path", "key"},
		},
		{
			name:      "empty key",
			key:       "",
			delimiter: ".",
			expected:  []string{""},
		},
		{
			name:      "single segment",
			key:       "single",
			delimiter: ".",
			expected:  []string{"single"},
		},
		{
			name:      "different delimiter",
			key:       "a/b/c",
			delimiter: "/",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "complex placeholder patterns",
			key:       "{abc.def-ghi_123}.{xyz-456.test}",
			delimiter: ".",
			expected:  []string{"abc.def-ghi_123", "xyz-456.test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := genPath(tt.key, tt.delimiter)
			assert.True(t, reflect.DeepEqual(tt.expected, result),
				"Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestGenPathEdgeCases(t *testing.T) {
	// Test regex compilation
	// This tests the global regex variable initialization
	assert.NotNil(t, regx)

	// Test malformed placeholder patterns
	tests := []struct {
		name      string
		key       string
		delimiter string
		expected  []string
	}{
		{
			name:      "unclosed placeholder",
			key:       "service.{unclosed.config",
			delimiter: ".",
			expected:  []string{"service", "{unclosed", "config"},
		},
		{
			name:      "consecutive placeholders",
			key:       "{first}.{second}",
			delimiter: ".",
			expected:  []string{"first", "second"},
		},
		{
			name:      "placeholder at start",
			key:       "{start}.middle.end",
			delimiter: ".",
			expected:  []string{"start", "middle", "end"},
		},
		{
			name:      "placeholder at end",
			key:       "start.middle.{end}",
			delimiter: ".",
			expected:  []string{"start", "middle", "end"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := genPath(tt.key, tt.delimiter)
			assert.True(t, reflect.DeepEqual(tt.expected, result),
				"Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestKeyConstants(t *testing.T) {
	// Test that all key constants are properly formatted
	assert.Equal(t, "yggdrasil", KeyBase)
	assert.Equal(t, "resolver", KeySingleResolver)
	assert.Equal(t, "balancer", KeySingleBalancer)
	assert.Equal(t, "endpoints", KeySingleEndpoints)
	assert.Equal(t, "address", KeySingleAddress)
	assert.Equal(t, "protocol", KeySingleProtocol)
	assert.Equal(t, "metadata", KeySingleMetadata)

	// Test client keys
	assert.Equal(t, "yggdrasil.client", KeyClient)
	assert.Equal(t, "yggdrasil.client.{%s}", KeyClientInstance)
	assert.Equal(t, "yggdrasil.client.{%s}.endpoints", KeyClientEndpoints)
	assert.Equal(t, "yggdrasil.client.{%s}.namespace", KeyClientNamespace)
	assert.Equal(t, "yggdrasil.client.{%s}.protocolConfig.{%s}", KeyClientProtocolCfg)
	assert.Equal(t, "yggdrasil.client.{%s}.balancerConfig.{%s}", KeyClientBalancerCfg)
	assert.Equal(t, "yggdrasil.client.{%s}.interceptor", KeyClientInterceptor)
	assert.Equal(t, "yggdrasil.client.{%s}.interceptor.unary", KeyClientUnaryInt)
	assert.Equal(t, "yggdrasil.client.{%s}.interceptor.stream", KeyClientStreamInt)
	assert.Equal(t, "yggdrasil.client.{%s}.interceptor.config.{%s}", KeyClientIntCfg)

	// Test server keys
	assert.Equal(t, "yggdrasil.server", KeyServer)
	assert.Equal(t, "yggdrasil.server.protocol", KeyServerProtocol)

	// Test other service keys
	assert.Equal(t, "yggdrasil.governor", KeyGovernor)
	assert.Equal(t, "yggdrasil.rest", KeyRest)
	assert.Equal(t, "yggdrasil.rest.enable", KeyRestEnable)
	assert.Equal(t, "yggdrasil.rest.marshaler", KeyRestMarshaler)
	assert.Equal(t, "yggdrasil.rest.marshaler.support", KeyRestMarshalerSupport)
	assert.Equal(t, "yggdrasil.rest.marshaler.config.{%s}", KeyRestMarshalerCfg)

	// Test interceptor keys
	assert.Equal(t, "yggdrasil.interceptor", KeyInterceptor)
	assert.Equal(t, "yggdrasil.interceptor.unaryClient", KeyIntUnaryClient)
	assert.Equal(t, "yggdrasil.interceptor.streamClient", KeyIntStreamClient)
	assert.Equal(t, "yggdrasil.interceptor.unaryServer", KeyIntUnaryServe)
	assert.Equal(t, "yggdrasil.interceptor.streamServer", KeyIntStreamServer)
	assert.Equal(t, "yggdrasil.interceptor.config.{%s}", KeyInterceptorCfg)

	// Test remote and logger keys
	assert.Equal(t, "yggdrasil.remote.protocol.{%s}", KeyRemoteProto)
	assert.Equal(t, "yggdrasil.logger.Logger.level", KeyRemoteLgLevel)

	// Test application keys
	assert.Equal(t, "yggdrasil.application", KeyApplication)
	assert.Equal(t, "yggdrasil.application.name", KeyAppName)
	assert.Equal(t, "yggdrasil.application.region", KeyAppRegion)
	assert.Equal(t, "yggdrasil.application.zone", KeyAppZone)
	assert.Equal(t, "yggdrasil.application.campus", KeyAppCampus)
	assert.Equal(t, "yggdrasil.application.namespace", KeyAppNamespace)
	assert.Equal(t, "yggdrasil.application.version", KeyAppVersion)
	assert.Equal(t, "yggdrasil.application.metadata", KeyAppMetadata)

	// Test monitoring keys
	assert.Equal(t, "yggdrasil.tracer", KeyTracer)
	assert.Equal(t, "yggdrasil.meter", KeyMeter)
	assert.Equal(t, "yggdrasil.registry", KeyRegistry)

	// Test logger keys
	assert.Equal(t, "yggdrasil.logger", KeyLogger)
	assert.Equal(t, "yggdrasil.logger.level", KeyLoggerLevel)
	assert.Equal(t, "yggdrasil.logger.writer", KeyLoggerWriter)
	assert.Equal(t, "yggdrasil.logger.timeEncoder", KeyLoggerTimeEnc)
	assert.Equal(t, "yggdrasil.logger.durationEncoder", KeyLoggerDurEnc)
	assert.Equal(t, "yggdrasil.logger.printStack", KeyLoggerStack)

	// Test stats keys
	assert.Equal(t, "yggdrasil.stats", KeyStats)
	assert.Equal(t, "yggdrasil.stats.config", KeyStatsCfg)
}

func TestKeyConstantsWithJoin(t *testing.T) {
	// Test that the constants work correctly with the Join function
	assert.Equal(t, "yggdrasil.client.test", Join(KeyClient, "test"))
	assert.Equal(t, "yggdrasil.server.protocol.test", Join(KeyServerProtocol, "test"))
	assert.Equal(t, "yggdrasil.application.name.test", Join(KeyAppName, "test"))
	assert.Equal(t, "yggdrasil.logger.level.test", Join(KeyLoggerLevel, "test"))
}

func TestGenPathWithRealWorldKeys(t *testing.T) {
	// Test with actual keys that might be used in the application
	realWorldKeys := []struct {
		key      string
		expected []string
	}{
		{
			key:      "yggdrasil.client.{my-service}.endpoints",
			expected: []string{"yggdrasil", "client", "my-service", "endpoints"},
		},
		{
			key:      "yggdrasil.client.{service.namespace}.protocolConfig.{grpc}",
			expected: []string{"yggdrasil", "client", "service.namespace", "protocolConfig", "grpc"},
		},
		{
			key:      "yggdrasil.interceptor.config.{logging.level}",
			expected: []string{"yggdrasil", "interceptor", "config", "logging.level"},
		},
		{
			key:      "yggdrasil.remote.protocol.{xds}",
			expected: []string{"yggdrasil", "remote", "protocol", "xds"},
		},
		{
			key:      "database.{primary}.connection.timeout",
			expected: []string{"database", "primary", "connection", "timeout"},
		},
	}

	for _, test := range realWorldKeys {
		t.Run("key: "+test.key, func(t *testing.T) {
			result := genPath(test.key, ".")
			assert.True(t, reflect.DeepEqual(test.expected, result),
				"For key '%s': expected %v, got %v", test.key, test.expected, result)
		})
	}
}

func TestGenPathWithSpecialCharacters(t *testing.T) {
	// Test various special characters that might appear in placeholders
	specialCharTests := []struct {
		key      string
		expected []string
	}{
		{
			key:      "service.{name_with_underscores}.config",
			expected: []string{"service", "name_with_underscores", "config"},
		},
		{
			key:      "service.{name-with-dashes}.config",
			expected: []string{"service", "name-with-dashes", "config"},
		},
		{
			key:      "service.{name123with456numbers}.config",
			expected: []string{"service", "name123with456numbers", "config"},
		},
		{
			key:      "service.{name.multiple.dots}.config",
			expected: []string{"service", "name.multiple.dots", "config"},
		},
		{
			key:      "service.{name_mixed-123.chars}.config",
			expected: []string{"service", "name_mixed-123.chars", "config"},
		},
	}

	for _, test := range specialCharTests {
		t.Run("special: "+test.key, func(t *testing.T) {
			result := genPath(test.key, ".")
			assert.True(t, reflect.DeepEqual(test.expected, result),
				"For key '%s': expected %v, got %v", test.key, test.expected, result)
		})
	}
}
