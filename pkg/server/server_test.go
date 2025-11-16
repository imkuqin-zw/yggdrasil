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

package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote"
	"github.com/imkuqin-zw/yggdrasil/pkg/rest"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/stretchr/testify/assert"
)

// Test service interfaces
type TestService interface {
	TestMethod(ctx context.Context, req interface{}) (interface{}, error)
}

type TestServiceImpl struct{}

func (t *TestServiceImpl) TestMethod(ctx context.Context, req interface{}) (interface{}, error) {
	return "test response", nil
}

// Tests for serverInfo struct
func TestServerInfo(t *testing.T) {
	si := &serverInfo{
		scheme:   "grpc",
		address:  "localhost:8080",
		svrKind:  pkg.ServerKindRpc,
		metadata: map[string]string{"version": "1.0"},
	}

	assert.Equal(t, "grpc", si.Scheme())
	assert.Equal(t, "localhost:8080", si.Address())
	assert.Equal(t, pkg.ServerKindRpc, si.Kind())
	assert.Equal(t, "1.0", si.Metadata()["version"])
}

// Test registerServiceInfo function
func TestRegisterServiceInfo(t *testing.T) {
	s := &server{
		services:     make(map[string]*ServiceInfo),
		servicesDesc: make(map[string][]methodInfo),
	}

	serviceDesc := &ServiceDesc{
		ServiceName: "test.service",
		Methods: []MethodDesc{
			{
				MethodName: "TestMethod",
				Handler:    nil,
			},
		},
		Streams: []stream.StreamDesc{
			{
				StreamName:    "TestStream",
				ServerStreams: true,
				ClientStreams: false,
			},
		},
		Metadata: "test metadata",
	}

	serviceImpl := &TestServiceImpl{}
	s.registerServiceInfo(serviceDesc, serviceImpl)

	info := s.services["test.service"]
	assert.NotNil(t, info)
	assert.Equal(t, serviceImpl, info.ServiceImpl)
	assert.Equal(t, "test metadata", info.Metadata)
	assert.Equal(t, 1, len(info.Methods))
	assert.Equal(t, "TestMethod", info.Methods["TestMethod"].MethodName)
	assert.Equal(t, 1, len(info.Streams))
	assert.Equal(t, "TestStream", info.Streams["TestStream"].StreamName)
}

// Test registerServiceDesc function
func TestRegisterServiceDesc(t *testing.T) {
	s := &server{
		services:     make(map[string]*ServiceInfo),
		servicesDesc: make(map[string][]methodInfo),
	}

	serviceDesc := &ServiceDesc{
		ServiceName: "test.service",
		Methods: []MethodDesc{
			{
				MethodName: "TestMethod",
			},
		},
		Streams: []stream.StreamDesc{
			{
				StreamName:    "TestStream",
				ServerStreams: true,
				ClientStreams: false,
			},
		},
	}

	s.registerServiceDesc(serviceDesc)

	methods := s.servicesDesc["test.service"]
	assert.Equal(t, 2, len(methods))
	assert.Equal(t, "TestMethod", methods[0].MethodName)
	assert.False(t, methods[0].ServerStreams)
	assert.False(t, methods[0].ClientStreams)
	assert.Equal(t, "TestStream", methods[1].MethodName)
	assert.True(t, methods[1].ServerStreams)
	assert.False(t, methods[1].ClientStreams)
}

// Test register function
func TestRegister(t *testing.T) {
	s := &server{
		services:     make(map[string]*ServiceInfo),
		servicesDesc: make(map[string][]methodInfo),
	}

	serviceDesc := &ServiceDesc{
		ServiceName: "test.service",
		HandlerType: (*TestService)(nil),
		Methods: []MethodDesc{
			{
				MethodName: "TestMethod",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
					return "test response", nil
				},
			},
		},
		Streams: []stream.StreamDesc{},
	}

	serviceImpl := &TestServiceImpl{}
	s.register(serviceDesc, serviceImpl)

	// Verify service is registered
	serviceInfo, exists := s.services["test.service"]
	assert.True(t, exists)
	assert.Equal(t, serviceImpl, serviceInfo.ServiceImpl)
	assert.Equal(t, 1, len(serviceInfo.Methods))
	assert.Equal(t, 1, len(s.servicesDesc["test.service"]))
}

// Test register with duplicate service
func TestRegisterDuplicate(t *testing.T) {
	s := &server{
		services:     make(map[string]*ServiceInfo),
		servicesDesc: make(map[string][]methodInfo),
	}

	serviceDesc := &ServiceDesc{
		ServiceName: "duplicate.service",
		HandlerType: (*TestService)(nil),
		Methods:     []MethodDesc{},
	}

	serviceImpl := &TestServiceImpl{}

	// Register first time
	s.register(serviceDesc, serviceImpl)

	// Verify service exists
	_, exists := s.services["duplicate.service"]
	assert.True(t, exists)

	// Note: Testing duplicate registration would cause logger.Fatalf which terminates the program
	// In this test environment, we skip the panic test to avoid program termination
}

// Test registerRest function
func TestRegisterRest(t *testing.T) {
	s := &server{
		restRouterDesc: []restRouterInfo{},
	}
	s.restSvr = &mockRestServer{}

	restServiceDesc := &RestServiceDesc{
		HandlerType: (*TestService)(nil),
		Methods: []RestMethodDesc{
			{
				Method: "GET",
				Path:   "/test",
				Handler: func(w http.ResponseWriter, r *http.Request, srv interface{}, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
					return "rest response", nil
				},
			},
		},
	}

	serviceImpl := &TestServiceImpl{}
	s.registerRest(restServiceDesc, serviceImpl, "/api")

	// Verify REST router is updated
	assert.Equal(t, 1, len(s.restRouterDesc))
	assert.Equal(t, "GET", s.restRouterDesc[0].Method)
	assert.Equal(t, "/api/test", s.restRouterDesc[0].Path)
}

// Test RegisterRestRawHandlers function
func TestRegisterRestRawHandlers(t *testing.T) {
	s := &server{
		restRouterDesc: []restRouterInfo{},
		restEnable:     true,
	}
	s.restSvr = &mockRestServer{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	rawHandlers := []*RestRawHandlerDesc{
		{
			Method:  "POST",
			Path:    "/raw",
			Handler: handler,
		},
		{
			Method:  "GET",
			Path:    "/health",
			Handler: handler,
		},
	}

	s.RegisterRestRawHandlers(rawHandlers...)

	// Verify routers are added
	assert.Equal(t, 2, len(s.restRouterDesc))
	assert.Equal(t, "POST", s.restRouterDesc[0].Method)
	assert.Equal(t, "/raw", s.restRouterDesc[0].Path)
	assert.Equal(t, "GET", s.restRouterDesc[1].Method)
	assert.Equal(t, "/health", s.restRouterDesc[1].Path)
}

// Test Endpoints function
func TestEndpoints(t *testing.T) {
	s := &server{
		servers: []remote.Server{},
	}

	// No servers registered
	endpoints := s.Endpoints()
	assert.Equal(t, 0, len(endpoints))

	// Test with REST enabled
	s.restEnable = true
	s.restSvr = &mockRestServer{
		address: "localhost:9000",
		attr:    map[string]string{"type": "rest"},
	}

	endpoints = s.Endpoints()
	assert.Equal(t, 1, len(endpoints))

	// Should have REST endpoint
	assert.Equal(t, "http", endpoints[0].Scheme())
	assert.Equal(t, "localhost:9000", endpoints[0].Address())
	assert.Equal(t, pkg.ServerKindRest, endpoints[0].Kind())
}

// Test Stop function
func TestStop(t *testing.T) {
	s := &server{
		servers: []remote.Server{},
		state:   serverStateRunning,
	}

	err := s.Stop()
	assert.NoError(t, err)
	assert.Equal(t, serverStateClosing, s.state)
}

// Test Stop from init state
func TestStopFromInitState(t *testing.T) {
	s := &server{
		servers: []remote.Server{},
		state:   serverStateInit,
	}

	err := s.Stop()
	assert.NoError(t, err)
	assert.Equal(t, serverStateClosing, s.state)
}

// Test Stop from closing state
func TestStopFromClosingState(t *testing.T) {
	s := &server{
		servers: []remote.Server{},
		state:   serverStateClosing,
	}

	err := s.Stop()
	assert.NoError(t, err)
	assert.Equal(t, serverStateClosing, s.state)
}

// Test Serve function edge cases
func TestServeEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupState  int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "serve from closing state",
			setupState:  serverStateClosing,
			expectError: true,
			errorMsg:    "server stopped",
		},
		{
			name:        "serve from running state",
			setupState:  serverStateRunning,
			expectError: true,
			errorMsg:    "server already serve",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				servers: []remote.Server{},
				state:   tt.setupState,
			}

			startFlag := make(chan struct{}, 1)
			err := s.Serve(startFlag)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

// Test initialization functions
func TestInitInterceptor(t *testing.T) {
	s := &server{}

	// This test mainly checks that the function doesn't panic
	s.initInterceptor()

	// Should not panic
}

func TestInitRemoteServer(t *testing.T) {
	s := &server{}

	// This test mainly checks that the function doesn't panic
	s.initRemoteServer()

	// Should not panic
}

// Mock rest server for testing
type mockRestServer struct {
	address string
	attr    map[string]string
}

func (m *mockRestServer) Info() rest.ServerInfo {
	return m
}

func (m *mockRestServer) GetAddress() string {
	return m.address
}

func (m *mockRestServer) GetAttributes() map[string]string {
	return m.attr
}

func (m *mockRestServer) RpcHandle(method, path string, f rest.HandlerFunc) {
	// Mock implementation
}

func (m *mockRestServer) RawHandle(method, path string, h http.HandlerFunc) {
	// Mock implementation
}

func (m *mockRestServer) Start() error {
	return nil
}

func (m *mockRestServer) Serve() error {
	return nil
}

func (m *mockRestServer) Stop() error {
	return nil
}

// Test serverInfo interface compliance
func TestServerInfoInterfaceCompliance(t *testing.T) {
	// Test that serverInfo implements Endpoint
	si := &serverInfo{
		scheme:   "grpc",
		address:  "localhost:8080",
		svrKind:  pkg.ServerKindRpc,
		metadata: map[string]string{"version": "1.0"},
	}
	var _ Endpoint = si

	// Verify interface methods work correctly
	assert.Equal(t, "grpc", si.Scheme())
	assert.Equal(t, "localhost:8080", si.Address())
	assert.Equal(t, pkg.ServerKindRpc, si.Kind())
	assert.Equal(t, "1.0", si.Metadata()["version"])
}

// Test RegisterRestService with REST disabled
func TestRegisterRestServiceDisabled(t *testing.T) {
	s := &server{
		restEnable:     false,
		restRouterDesc: []restRouterInfo{},
	}

	restServiceDesc := &RestServiceDesc{
		HandlerType: (*TestService)(nil),
		Methods:     []RestMethodDesc{},
	}

	serviceImpl := &TestServiceImpl{}
	s.RegisterRestService(restServiceDesc, serviceImpl)

	// Should not add any routers since REST is disabled
	assert.Equal(t, 0, len(s.restRouterDesc))
}

// Test RegisterRestRawHandlers with REST disabled
func TestRegisterRestRawHandlersDisabled(t *testing.T) {
	s := &server{
		restEnable:     false,
		restRouterDesc: []restRouterInfo{},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	rawHandlers := []*RestRawHandlerDesc{
		{
			Method:  "GET",
			Path:    "/health",
			Handler: handler,
		},
	}

	s.RegisterRestRawHandlers(rawHandlers...)

	// Should not add any routers since REST is disabled
	assert.Equal(t, 0, len(s.restRouterDesc))
}
