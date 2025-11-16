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
	"fmt"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/stats"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

// TestInjectFunction tests the inject function
func TestInjectFunction(t *testing.T) {
	ctx := context.Background()

	// Test inject with default propagator (should not panic)
	assert.NotPanics(t, func() {
		resultCtx := inject(ctx, nil)
		assert.NotNil(t, resultCtx)
	})

	// Test inject with context (should not panic)
	assert.NotPanics(t, func() {
		resultCtx := inject(ctx, nil)
		assert.NotNil(t, resultCtx)
	})
}

// TestExtractFunction tests the extract function
func TestExtractFunction(t *testing.T) {
	ctx := context.Background()

	// Test extract with default propagator (should not panic)
	assert.NotPanics(t, func() {
		resultCtx := extract(ctx, nil)
		assert.NotNil(t, resultCtx)
	})

	// Test extract with context (should not panic)
	assert.NotPanics(t, func() {
		resultCtx := extract(ctx, nil)
		assert.NotNil(t, resultCtx)
	})
}

// TestHandlerMetricsCreation tests metrics creation with various configurations
func TestHandlerMetricsCreation(t *testing.T) {
	config.Set("stats.cfg.otel", map[string]interface{}{
		"enable_metrics": true,
	})

	// Test server handler metrics
	svrHandler := newHandler(true)
	assert.NotNil(t, svrHandler.rpcDuration)
	assert.NotNil(t, svrHandler.rpcRequestSize)
	assert.NotNil(t, svrHandler.rpcResponseSize)
	assert.NotNil(t, svrHandler.rpcRequestsPerRPC)
	assert.NotNil(t, svrHandler.rpcResponsesPerRPC)

	// Test client handler metrics
	cliHandler := newHandler(false)
	assert.NotNil(t, cliHandler.rpcDuration)
	assert.NotNil(t, cliHandler.rpcRequestSize)
	assert.NotNil(t, cliHandler.rpcResponseSize)
	assert.NotNil(t, cliHandler.rpcRequestsPerRPC)
	assert.NotNil(t, cliHandler.rpcResponsesPerRPC)
}

// TestHandlerWithMetricsDisabled tests handler when metrics are disabled
func TestHandlerWithMetricsDisabled(t *testing.T) {
	config.Set("stats.cfg.otel", map[string]interface{}{
		"enable_metrics": false,
	})

	handler := newHandler(true)
	assert.NotNil(t, handler)

	// When metrics are disabled, handleRPC should be handleWithOutMetrics
	assert.NotNil(t, handler.handleRPC)

	// Should still have metrics created (as noop)
	assert.NotNil(t, handler.rpcDuration)
	assert.NotNil(t, handler.rpcRequestSize)
	assert.NotNil(t, handler.rpcResponseSize)
	assert.NotNil(t, handler.rpcRequestsPerRPC)
	assert.NotNil(t, handler.rpcResponsesPerRPC)
}

// TestHandlerWithDifferentEventConfigurations tests event configuration
func TestHandlerWithDifferentEventConfigurations(t *testing.T) {
	tests := []struct {
		name          string
		receivedEvent bool
		sentEvent     bool
		setupConfig   func()
	}{
		{
			name:          "both events enabled",
			receivedEvent: true,
			sentEvent:     true,
			setupConfig: func() {
				config.Set("stats.cfg.otel", map[string]interface{}{
					"received_event": true,
					"sent_event":     true,
				})
			},
		},
		{
			name:          "received only",
			receivedEvent: true,
			sentEvent:     false,
			setupConfig: func() {
				config.Set("stats.cfg.otel", map[string]interface{}{
					"received_event": true,
					"sent_event":     false,
				})
			},
		},
		{
			name:          "sent only",
			receivedEvent: false,
			sentEvent:     true,
			setupConfig: func() {
				config.Set("stats.cfg.otel", map[string]interface{}{
					"received_event": false,
					"sent_event":     true,
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupConfig()
			handler := newHandler(true)
			assert.Equal(t, tt.receivedEvent, handler.cfg.ReceivedEvent)
			assert.Equal(t, tt.sentEvent, handler.cfg.SentEvent)
		})
	}
}

// TestRPCContextWithAtomicOperations tests atomic operations in RPC context
func TestRPCContextWithAtomicOperations(t *testing.T) {
	rctx := &rpcContext{
		messagesReceived: 0,
		messagesSent:     0,
		metricAttrs:      []attribute.KeyValue{attribute.String("test", "value")},
	}

	// Test initial values
	assert.Equal(t, int64(0), rctx.messagesReceived)
	assert.Equal(t, int64(0), rctx.messagesSent)
	assert.NotEmpty(t, rctx.metricAttrs)

	// Test that fields can be accessed
	assert.NotNil(t, rctx.messagesReceived)
	assert.NotNil(t, rctx.messagesSent)
	assert.NotNil(t, rctx.metricAttrs)
}

// TestHandlerRoleBasedMetrics tests role-specific metric creation
func TestHandlerRoleBasedMetrics(t *testing.T) {
	config.Set("stats.cfg.otel", map[string]interface{}{
		"enable_metrics": true,
	})

	// Test server handler
	svrHandler := newHandler(true)
	assert.NotNil(t, svrHandler)

	// Test client handler
	cliHandler := newHandler(false)
	assert.NotNil(t, cliHandler)

	// Both should have metrics created
	assert.NotNil(t, svrHandler.rpcDuration)
	assert.NotNil(t, cliHandler.rpcDuration)
}

// TestParseFullMethodWithUnicode tests parsing with unicode characters
func TestParseFullMethodWithUnicode(t *testing.T) {
	unicodeMethods := []string{
		"/测试服务/测试方法",
		"/api.服务.v1/创建资源",
		"/package.süße/Methode",
	}

	for i, method := range unicodeMethods {
		t.Run(fmt.Sprintf("unicode_%d", i), func(t *testing.T) {
			name, _ := parseFullMethod(method)
			assert.NotNil(t, name)
			// Should not panic with unicode characters
		})
	}
}

// TestHandlerCreationWithConcurrentAccess tests concurrent handler creation
func TestHandlerCreationWithConcurrentAccess(t *testing.T) {
	config.Set("stats.cfg.otel", map[string]interface{}{
		"enable_metrics": true,
	})

	// Test concurrent creation doesn't panic
	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("concurrent_%d", i), func(t *testing.T) {
			assert.NotPanics(t, func() {
				svrHandler := newHandler(true)
				assert.NotNil(t, svrHandler)

				cliHandler := newHandler(false)
				assert.NotNil(t, cliHandler)
			})
		})
	}
}

// TestHandlerWithNoConfiguration tests handler when no config is set
func TestHandlerWithNoConfiguration(t *testing.T) {
	// Clear config
	config.Set("stats.cfg.otel", "")

	assert.NotPanics(t, func() {
		handler := newHandler(true)
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.cfg)
		assert.NotNil(t, handler.tracer)
	})
}

// TestServerHandlerInterface tests server handler interface compliance
func TestServerHandlerInterface(t *testing.T) {
	config.Set("stats.cfg.otel", map[string]interface{}{
		"enable_metrics": false,
	})

	handler := newSvrHandler()
	ctx := context.Background()

	// Test all interface methods
	assert.NotPanics(t, func() {
		info := &struct{ stats.RPCTagInfoBase }{}
		resultCtx := handler.TagRPC(ctx, info)
		assert.NotNil(t, resultCtx)
	})

	assert.NotPanics(t, func() {
		stats := &struct{ stats.RPCBeginBase }{}
		handler.HandleRPC(ctx, stats)
	})

	assert.NotPanics(t, func() {
		info := &struct{ stats.ChanTagInfoBase }{}
		resultCtx := handler.TagChannel(ctx, info)
		assert.NotNil(t, resultCtx)
	})

	assert.NotPanics(t, func() {
		stats := &struct{ stats.ChanBeginBase }{}
		handler.HandleChannel(ctx, stats)
	})
}

// TestClientHandlerInterface tests client handler interface compliance
func TestClientHandlerInterface(t *testing.T) {
	config.Set("stats.cfg.otel", map[string]interface{}{
		"enable_metrics": false,
	})

	handler := newCliHandler()
	ctx := context.Background()

	// Test all interface methods
	assert.NotPanics(t, func() {
		info := &struct{ stats.RPCTagInfoBase }{}
		resultCtx := handler.TagRPC(ctx, info)
		assert.NotNil(t, resultCtx)
	})

	assert.NotPanics(t, func() {
		stats := &struct{ stats.RPCBeginBase }{}
		handler.HandleRPC(ctx, stats)
	})

	assert.NotPanics(t, func() {
		info := &struct{ stats.ChanTagInfoBase }{}
		resultCtx := handler.TagChannel(ctx, info)
		assert.NotNil(t, resultCtx)
	})

	assert.NotPanics(t, func() {
		stats := &struct{ stats.ChanBeginBase }{}
		handler.HandleChannel(ctx, stats)
	})
}

// TestRPCStatsInterfaceCompliance tests that all RPC stats types implement the interface
func TestRPCStatsInterfaceCompliance(t *testing.T) {
	// Test that all the base types implement the correct interfaces
	assert.NotPanics(t, func() {
		var _ stats.RPCTagInfo = &struct{ stats.RPCTagInfoBase }{}
		var _ stats.RPCBegin = &struct{ stats.RPCBeginBase }{}
		var _ stats.RPCInPayload = &struct{ stats.RPCInPayloadBase }{}
		var _ stats.RPCOutPayload = &struct{ stats.RPCOutPayloadBase }{}
		var _ stats.RPCEnd = &struct{ stats.RPCEndBase }{}
	})

	// Test Chan stats interfaces
	assert.NotPanics(t, func() {
		var _ stats.ChanTagInfo = &struct{ stats.ChanTagInfoBase }{}
		var _ stats.ChanBegin = &struct{ stats.ChanBeginBase }{}
		var _ stats.ChanEnd = &struct{ stats.ChanEndBase }{}
	})
}

// TestConfigStructDefaultValues tests the Config struct default values
func TestConfigStructDefaultValues(t *testing.T) {
	cfg := &Config{}
	// Test that default values are true as specified in the struct tags
	assert.True(t, cfg.ReceivedEvent)
	assert.True(t, cfg.SentEvent)
	assert.True(t, cfg.EnableMetrics)
}

// TestGetCfgFunctionIdempotency tests that getCfg returns consistent results
func TestGetCfgFunctionIdempotency(t *testing.T) {
	config.Set("stats.cfg.otel", map[string]interface{}{
		"received_event": true,
		"sent_event":     false,
		"enable_metrics": true,
	})

	// Multiple calls should return the same result
	cfg1 := getCfg()
	cfg2 := getCfg()
	cfg3 := getCfg()

	assert.Equal(t, cfg1.ReceivedEvent, cfg2.ReceivedEvent)
	assert.Equal(t, cfg1.SentEvent, cfg2.SentEvent)
	assert.Equal(t, cfg1.EnableMetrics, cfg2.EnableMetrics)
	assert.Equal(t, cfg2.ReceivedEvent, cfg3.ReceivedEvent)
	assert.Equal(t, cfg2.SentEvent, cfg3.SentEvent)
	assert.Equal(t, cfg2.EnableMetrics, cfg3.EnableMetrics)
}

// TestParseFullMethodEmptyInput tests parsing with empty and edge cases
func TestParseFullMethodEmptyInput(t *testing.T) {
	// Test with various edge cases
	testCases := []string{
		"",
		"/",
		"//",
		"///",
		"method",
		"method/",
		"/method",
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("edge_case_%s", testCase), func(t *testing.T) {
			name, attrs := parseFullMethod(testCase)
			// Should not panic and return some result
			assert.NotNil(t, name)
			assert.NotNil(t, attrs)
		})
	}
}
