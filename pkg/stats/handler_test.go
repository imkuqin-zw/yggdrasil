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

package stats

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for testing

type MockHandler struct {
	tagRPCCalled     bool
	handleRPCCalled  bool
	tagChanCalled    bool
	handleChanCalled bool
	rpcInfo          RPCTagInfo
	chanInfo         ChanTagInfo
	rpcStats         RPCStats
	chanStats        ChanStats
}

func NewMockHandler() *MockHandler {
	return &MockHandler{}
}

func (m *MockHandler) TagRPC(ctx context.Context, info RPCTagInfo) context.Context {
	m.tagRPCCalled = true
	m.rpcInfo = info
	return ctx
}

func (m *MockHandler) HandleRPC(ctx context.Context, rs RPCStats) {
	m.handleRPCCalled = true
	m.rpcStats = rs
}

func (m *MockHandler) TagChannel(ctx context.Context, info ChanTagInfo) context.Context {
	m.tagChanCalled = true
	m.chanInfo = info
	return ctx
}

func (m *MockHandler) HandleChannel(ctx context.Context, cs ChanStats) {
	m.handleChanCalled = true
	m.chanStats = cs
}

func (m *MockHandler) Reset() {
	m.tagRPCCalled = false
	m.handleRPCCalled = false
	m.tagChanCalled = false
	m.handleChanCalled = false
}

type MockRPCTagInfo struct {
	fullMethod string
}

func (m *MockRPCTagInfo) GetFullMethod() string {
	return m.fullMethod
}

func (m *MockRPCTagInfo) isRPCTagInfo() {}

type MockChanTagInfo struct {
	protocol       string
	remoteEndpoint string
	localEndpoint  string
}

func (m *MockChanTagInfo) GetProtocol() string {
	return m.protocol
}

func (m *MockChanTagInfo) GetRemoteEndpoint() string {
	return m.remoteEndpoint
}

func (m *MockChanTagInfo) GetLocalEndpoint() string {
	return m.localEndpoint
}

func (m *MockChanTagInfo) isChanTagInfo() {}

type MockRPCStats struct{}

func (m *MockRPCStats) isRPCStats() {}

type MockChanStats struct {
	client bool
}

func (m *MockChanStats) isChanStats() {}

func (m *MockChanStats) IsClient() bool {
	return m.client
}

// Test RegisterHandlerBuilder function
func TestRegisterHandlerBuilder(t *testing.T) {
	// Clear global state before test
	clearHandlerBuilders()

	tests := []struct {
		name         string
		handlerName  string
		builder      HandlerBuilder
		expectResult HandlerBuilder
	}{
		{
			name:        "register new handler builder",
			handlerName: "test-handler",
			builder: func(isServer bool) Handler {
				return NewMockHandler()
			},
			expectResult: func(isServer bool) Handler {
				return NewMockHandler()
			},
		},
		{
			name:         "register nil handler builder",
			handlerName:  "nil-handler",
			builder:      nil,
			expectResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearHandlerBuilders()

			RegisterHandlerBuilder(tt.handlerName, tt.builder)

			// Verify the builder is registered
			retrievedBuilder := GetHandlerBuilder(tt.handlerName)
			if tt.expectResult == nil {
				assert.Nil(t, retrievedBuilder)
			} else {
				assert.NotNil(t, retrievedBuilder)
			}
		})
	}
}

// Test GetHandlerBuilder function
func TestGetHandlerBuilder(t *testing.T) {
	// Clear global state before test
	clearHandlerBuilders()

	tests := []struct {
		name        string
		setup       func()
		handlerName string
		expectNil   bool
	}{
		{
			name: "get existing handler builder",
			setup: func() {
				RegisterHandlerBuilder("existing", func(isServer bool) Handler {
					return NewMockHandler()
				})
			},
			handlerName: "existing",
			expectNil:   false,
		},
		{
			name:        "get non-existent handler builder",
			setup:       func() {},
			handlerName: "non-existent",
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearHandlerBuilders()
			tt.setup()

			builder := GetHandlerBuilder(tt.handlerName)

			if tt.expectNil {
				assert.Nil(t, builder)
			} else {
				assert.NotNil(t, builder)
			}
		})
	}
}

// Test handlerChain functionality
func TestHandlerChain(t *testing.T) {
	handler1 := NewMockHandler()
	handler2 := NewMockHandler()
	handler3 := NewMockHandler()

	chain := &handlerChain{
		handlers: []Handler{handler1, handler2, handler3},
	}

	// Test TagRPC
	ctx := context.Background()
	rpcInfo := &MockRPCTagInfo{fullMethod: "/test.service/method"}
	newCtx := chain.TagRPC(ctx, rpcInfo)

	assert.Equal(t, ctx, newCtx) // Context should be returned (modified or not)
	assert.True(t, handler1.tagRPCCalled)
	assert.True(t, handler2.tagRPCCalled)
	assert.True(t, handler3.tagRPCCalled)

	// Reset handlers
	handler1.Reset()
	handler2.Reset()
	handler3.Reset()

	// Test HandleRPC
	rpcStats := &MockRPCStats{}
	chain.HandleRPC(ctx, rpcStats)

	assert.True(t, handler1.handleRPCCalled)
	assert.True(t, handler2.handleRPCCalled)
	assert.True(t, handler3.handleRPCCalled)
	assert.Equal(t, rpcStats, handler1.rpcStats)

	// Reset handlers
	handler1.Reset()
	handler2.Reset()
	handler3.Reset()

	// Test TagChannel
	chanInfo := &MockChanTagInfo{
		protocol:       "grpc",
		remoteEndpoint: "localhost:8080",
		localEndpoint:  "localhost:9090",
	}
	newCtx = chain.TagChannel(ctx, chanInfo)

	assert.Equal(t, ctx, newCtx)
	assert.True(t, handler1.tagChanCalled)
	assert.True(t, handler2.tagChanCalled)
	assert.True(t, handler3.tagChanCalled)

	// Reset handlers
	handler1.Reset()
	handler2.Reset()
	handler3.Reset()

	// Test HandleChannel
	chanStats := &MockChanStats{client: true}
	chain.HandleChannel(ctx, chanStats)

	assert.True(t, handler1.handleChanCalled)
	assert.True(t, handler2.handleChanCalled)
	assert.True(t, handler3.handleChanCalled)
	assert.Equal(t, chanStats, handler1.chanStats)
}

// Test handlerChain with empty handlers
func TestHandlerChainEmpty(t *testing.T) {
	chain := &handlerChain{
		handlers: []Handler{},
	}

	ctx := context.Background()
	rpcInfo := &MockRPCTagInfo{fullMethod: "/test.service/method"}
	chanInfo := &MockChanTagInfo{protocol: "grpc"}
	rpcStats := &MockRPCStats{}
	chanStats := &MockChanStats{client: true}

	// All methods should work without panic
	assert.NotPanics(t, func() {
		chain.TagRPC(ctx, rpcInfo)
	})
	assert.NotPanics(t, func() {
		chain.HandleRPC(ctx, rpcStats)
	})
	assert.NotPanics(t, func() {
		chain.TagChannel(ctx, chanInfo)
	})
	assert.NotPanics(t, func() {
		chain.HandleChannel(ctx, chanStats)
	})
}

// Test GetServerHandler function
func TestGetServerHandler(t *testing.T) {
	// Clear global state before test
	clearGlobalHandlers()

	// Reset config to empty
	config.Set(config.Join(config.KeyStats, "server"), "")

	handler := GetServerHandler()
	assert.NotNil(t, handler)

	// Test singleton behavior - should return the same instance
	handler2 := GetServerHandler()
	assert.Equal(t, handler, handler2)
}

// Test GetClientHandler function
func TestGetClientHandler(t *testing.T) {
	// Clear global state before test
	clearGlobalHandlers()

	// Reset config to empty
	config.Set(config.Join(config.KeyStats, "client"), "")

	handler := GetClientHandler()
	assert.NotNil(t, handler)

	// Test singleton behavior - should return the same instance
	handler2 := GetClientHandler()
	assert.Equal(t, handler, handler2)
}

// Test GetServerHandler with configuration
func TestGetServerHandlerWithConfig(t *testing.T) {
	// Clear global state before test
	clearHandlerBuilders()
	clearGlobalHandlers()

	// Register a test handler builder
	RegisterHandlerBuilder("test-handler", func(isServer bool) Handler {
		handler := NewMockHandler()
		// Set a flag to verify isServer parameter
		if isServer {
			return &ServerMarkedHandler{MockHandler: handler}
		}
		return handler
	})

	// Set config to use the test handler
	config.Set(config.Join(config.KeyStats, "server"), "test-handler")

	handler := GetServerHandler()
	assert.NotNil(t, handler)

	// GetServerHandler returns a handlerChain, check if it contains our server-marked handler
	chain, ok := handler.(*handlerChain)
	assert.True(t, ok)
	assert.Len(t, chain.handlers, 1)

	_, isServerMarked := chain.handlers[0].(*ServerMarkedHandler)
	assert.True(t, isServerMarked)

	// Test singleton behavior
	handler2 := GetServerHandler()
	assert.Equal(t, handler, handler2)
}

// Test GetClientHandler with configuration
func TestGetClientHandlerWithConfig(t *testing.T) {
	// Clear global state before test
	clearHandlerBuilders()
	clearGlobalHandlers()

	// Register a test handler builder
	RegisterHandlerBuilder("test-handler", func(isServer bool) Handler {
		handler := NewMockHandler()
		// Set a flag to verify isServer parameter
		if !isServer {
			return &ClientMarkedHandler{MockHandler: handler}
		}
		return handler
	})

	// Set config to use the test handler
	config.Set(config.Join(config.KeyStats, "client"), "test-handler")

	handler := GetClientHandler()
	assert.NotNil(t, handler)

	// GetClientHandler returns a handlerChain, check if it contains our client-marked handler
	chain, ok := handler.(*handlerChain)
	assert.True(t, ok)
	assert.Len(t, chain.handlers, 1)

	_, isClientMarked := chain.handlers[0].(*ClientMarkedHandler)
	assert.True(t, isClientMarked)

	// Test singleton behavior
	handler2 := GetClientHandler()
	assert.Equal(t, handler, handler2)
}

// Test handler configuration parsing
func TestHandlerConfigParsing(t *testing.T) {
	// Clear global state before test
	clearHandlerBuilders()
	clearGlobalHandlers()

	// Register test handlers
	RegisterHandlerBuilder("handler1", func(isServer bool) Handler {
		return &TestNamedHandler{name: "handler1", isServer: isServer}
	})
	RegisterHandlerBuilder("handler2", func(isServer bool) Handler {
		return &TestNamedHandler{name: "handler2", isServer: isServer}
	})

	tests := []struct {
		name      string
		config    string
		expectNum int
		isServer  bool
	}{
		{
			name:      "single handler",
			config:    "handler1",
			expectNum: 1,
			isServer:  true,
		},
		{
			name:      "multiple handlers",
			config:    "handler1,handler2",
			expectNum: 2,
			isServer:  true,
		},
		{
			name:      "with empty entries",
			config:    "handler1,,handler2,",
			expectNum: 2,
			isServer:  true,
		},
		{
			name:      "empty config",
			config:    "",
			expectNum: 0,
			isServer:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearGlobalHandlers()

			if tt.isServer {
				config.Set(config.Join(config.KeyStats, "server"), tt.config)
				handler := GetServerHandler()
				chain, ok := handler.(*handlerChain)
				if tt.expectNum > 0 {
					assert.True(t, ok)
					assert.Equal(t, tt.expectNum, len(chain.handlers))
				} else {
					assert.True(t, ok)
					assert.Equal(t, 0, len(chain.handlers))
				}
			} else {
				config.Set(config.Join(config.KeyStats, "client"), tt.config)
				handler := GetClientHandler()
				chain, ok := handler.(*handlerChain)
				if tt.expectNum > 0 {
					assert.True(t, ok)
					assert.Equal(t, tt.expectNum, len(chain.handlers))
				} else {
					assert.True(t, ok)
					assert.Equal(t, 0, len(chain.handlers))
				}
			}
		})
	}
}

// Test concurrent handler registration
func TestConcurrentHandlerRegistration(t *testing.T) {
	// Clear global state before test
	clearHandlerBuilders()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			handlerName := fmt.Sprintf("handler-%d", index)
			RegisterHandlerBuilder(handlerName, func(isServer bool) Handler {
				return &TestNamedHandler{name: handlerName, isServer: isServer}
			})
		}(i)
	}

	wg.Wait()

	// Verify all handlers were registered
	for i := 0; i < numGoroutines; i++ {
		handlerName := fmt.Sprintf("handler-%d", i)
		builder := GetHandlerBuilder(handlerName)
		assert.NotNil(t, builder, "Handler %s should be registered", handlerName)
	}
}

// Test concurrent handler access
func TestConcurrentHandlerAccess(t *testing.T) {
	// Clear global state before test
	clearHandlerBuilders()
	clearGlobalHandlers()

	// Register a test handler
	RegisterHandlerBuilder("test-handler", func(isServer bool) Handler {
		return NewMockHandler()
	})

	// Set config
	config.Set(config.Join(config.KeyStats, "server"), "test-handler")

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent access to GetServerHandler
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler := GetServerHandler()
			assert.NotNil(t, handler)
		}()
	}

	wg.Wait()
}

// Helper types for testing

type ServerMarkedHandler struct {
	*MockHandler
}

type ClientMarkedHandler struct {
	*MockHandler
}

type TestNamedHandler struct {
	name     string
	isServer bool
}

func (t *TestNamedHandler) TagRPC(ctx context.Context, info RPCTagInfo) context.Context {
	return ctx
}

func (t *TestNamedHandler) HandleRPC(ctx context.Context, rs RPCStats) {
}

func (t *TestNamedHandler) TagChannel(ctx context.Context, info ChanTagInfo) context.Context {
	return ctx
}

func (t *TestNamedHandler) HandleChannel(ctx context.Context, cs ChanStats) {
}

// Helper functions to reset global state

func clearHandlerBuilders() {
	mu.Lock()
	handlerBuilder = make(map[string]HandlerBuilder)
	mu.Unlock()
}

func clearGlobalHandlers() {
	svrOnce = sync.Once{}
	svrHandler = nil
	cliOnce = sync.Once{}
	cliHandler = nil
}

// Test that handlerChain satisfies Handler interface
func TestHandlerChainInterface(t *testing.T) {
	var _ Handler = (*handlerChain)(nil)

	chain := &handlerChain{handlers: []Handler{NewMockHandler()}}
	assert.NotNil(t, chain)
}
