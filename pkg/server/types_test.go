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
	"fmt"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/stretchr/testify/assert"
)

// Test implementations for interfaces

type MockEndpoint struct {
	scheme   string
	address  string
	metadata map[string]string
	kind     pkg.ServerKind
}

func NewMockEndpoint(scheme, address string, kind pkg.ServerKind) *MockEndpoint {
	return &MockEndpoint{
		scheme:  scheme,
		address: address,
		kind:    kind,
		metadata: map[string]string{
			"version": "1.0",
			"region":  "us-east-1",
		},
	}
}

func (m *MockEndpoint) Scheme() string {
	return m.scheme
}

func (m *MockEndpoint) Address() string {
	return m.address
}

func (m *MockEndpoint) Metadata() map[string]string {
	return m.metadata
}

func (m *MockEndpoint) Kind() pkg.ServerKind {
	return m.kind
}

type MockServer struct {
	services          []*ServiceDesc
	restServices      []*RestServiceDesc
	rawHandlers       []*RestRawHandlerDesc
	serveCalled       bool
	stopCalled        bool
	endpoints         []Endpoint
	startFlagProvided bool
}

func NewMockServer() *MockServer {
	return &MockServer{
		services:     make([]*ServiceDesc, 0),
		restServices: make([]*RestServiceDesc, 0),
		rawHandlers:  make([]*RestRawHandlerDesc, 0),
		endpoints:    make([]Endpoint, 0),
	}
}

func (m *MockServer) RegisterService(sd *ServiceDesc, ss interface{}) {
	m.services = append(m.services, sd)
}

func (m *MockServer) RegisterRestService(sd *RestServiceDesc, ss interface{}, prefix ...string) {
	m.restServices = append(m.restServices, sd)
}

func (m *MockServer) RegisterRestRawHandlers(sd ...*RestRawHandlerDesc) {
	m.rawHandlers = append(m.rawHandlers, sd...)
}

func (m *MockServer) Serve(startFlag chan<- struct{}) error {
	m.serveCalled = true
	if startFlag != nil {
		startFlag <- struct{}{}
		m.startFlagProvided = true
	}
	return nil
}

func (m *MockServer) Stop() error {
	m.stopCalled = true
	return nil
}

func (m *MockServer) Endpoints() []Endpoint {
	return m.endpoints
}

// Test Endpoint interface
func TestEndpointInterface(t *testing.T) {
	// Test RPC endpoint
	rpcEndpoint := NewMockEndpoint("grpc", "localhost:8080", pkg.ServerKindRpc)

	assert.Equal(t, "grpc", rpcEndpoint.Scheme())
	assert.Equal(t, "localhost:8080", rpcEndpoint.Address())
	assert.Equal(t, pkg.ServerKindRpc, rpcEndpoint.Kind())

	metadata := rpcEndpoint.Metadata()
	assert.Equal(t, "1.0", metadata["version"])
	assert.Equal(t, "us-east-1", metadata["region"])

	// Test REST endpoint
	restEndpoint := NewMockEndpoint("http", "localhost:9000", pkg.ServerKindRest)

	assert.Equal(t, "http", restEndpoint.Scheme())
	assert.Equal(t, "localhost:9000", restEndpoint.Address())
	assert.Equal(t, pkg.ServerKindRest, restEndpoint.Kind())

	metadata = restEndpoint.Metadata()
	assert.Equal(t, "1.0", metadata["version"])
	assert.Equal(t, "us-east-1", metadata["region"])
}

// Test Server interface
func TestServerInterface(t *testing.T) {
	server := NewMockServer()

	// Test initial state
	assert.False(t, server.serveCalled)
	assert.False(t, server.stopCalled)
	assert.False(t, server.startFlagProvided)
	assert.Equal(t, 0, len(server.services))
	assert.Equal(t, 0, len(server.restServices))
	assert.Equal(t, 0, len(server.rawHandlers))

	// Test RegisterService
	serviceDesc := &ServiceDesc{
		ServiceName: "test.service",
		HandlerType: (*TestService)(nil),
		Methods:     []MethodDesc{},
	}

	server.RegisterService(serviceDesc, nil)
	assert.Equal(t, 1, len(server.services))
	assert.Equal(t, serviceDesc, server.services[0])

	// Test RegisterRestService
	restServiceDesc := &RestServiceDesc{
		HandlerType: (*TestService)(nil),
		Methods:     []RestMethodDesc{},
	}

	server.RegisterRestService(restServiceDesc, nil)
	assert.Equal(t, 1, len(server.restServices))
	assert.Equal(t, restServiceDesc, server.restServices[0])

	// Test RegisterRestRawHandlers
	rawHandler := &RestRawHandlerDesc{
		Method:  "GET",
		Path:    "/health",
		Handler: nil, // Not testing handler functionality here
	}

	server.RegisterRestRawHandlers(rawHandler)
	assert.Equal(t, 1, len(server.rawHandlers))
	assert.Equal(t, rawHandler, server.rawHandlers[0])

	// Test Serve
	startFlag := make(chan struct{}, 1)
	err := server.Serve(startFlag)
	assert.NoError(t, err)
	assert.True(t, server.serveCalled)
	assert.True(t, server.startFlagProvided)

	// Verify startFlag was signaled
	select {
	case <-startFlag:
		// Expected
	default:
		t.Error("Start flag was not signaled")
	}

	// Test Endpoints
	endpoint := NewMockEndpoint("grpc", "localhost:8080", pkg.ServerKindRpc)
	server.endpoints = append(server.endpoints, endpoint)

	endpoints := server.Endpoints()
	assert.Equal(t, 1, len(endpoints))
	assert.Equal(t, endpoint, endpoints[0])

	// Test Stop
	err = server.Stop()
	assert.NoError(t, err)
	assert.True(t, server.stopCalled)
}

// Test interface implementations
func TestInterfaceImplementations(t *testing.T) {
	// Test that MockEndpoint implements Endpoint
	var _ Endpoint = (*MockEndpoint)(nil)

	// Test that MockServer implements Server
	var _ Server = (*MockServer)(nil)

	// Verify interface methods work correctly
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

// Test serverInfo implementation
func TestServerInfoImplementation(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		address  string
		kind     pkg.ServerKind
		metadata map[string]string
	}{
		{
			name:    "gRPC endpoint",
			scheme:  "grpc",
			address: "localhost:8080",
			kind:    pkg.ServerKindRpc,
			metadata: map[string]string{
				"protocol": "grpc",
				"version":  "v1",
			},
		},
		{
			name:    "HTTP endpoint",
			scheme:  "http",
			address: "localhost:9000",
			kind:    pkg.ServerKindRest,
			metadata: map[string]string{
				"protocol": "http",
				"version":  "v1",
			},
		},
		{
			name:     "endpoint with empty metadata",
			scheme:   "https",
			address:  "api.example.com:443",
			kind:     pkg.ServerKindRpc,
			metadata: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &serverInfo{
				scheme:   tt.scheme,
				address:  tt.address,
				svrKind:  tt.kind,
				metadata: tt.metadata,
			}

			// Test all interface methods
			assert.Equal(t, tt.scheme, si.Scheme())
			assert.Equal(t, tt.address, si.Address())
			assert.Equal(t, tt.kind, si.Kind())
			assert.Equal(t, tt.metadata, si.Metadata())

			// Test specific metadata values
			for key, value := range tt.metadata {
				assert.Equal(t, value, si.Metadata()[key])
			}
		})
	}
}

// Test Server interface methods with edge cases
func TestServerInterfaceEdgeCases(t *testing.T) {
	server := NewMockServer()

	// Test multiple service registrations
	for i := 0; i < 5; i++ {
		serviceDesc := &ServiceDesc{
			ServiceName: fmt.Sprintf("service.%d", i),
			HandlerType: (*TestService)(nil),
			Methods:     []MethodDesc{},
		}
		server.RegisterService(serviceDesc, nil)
	}
	assert.Equal(t, 5, len(server.services))

	// Test multiple REST service registrations
	for i := 0; i < 3; i++ {
		restServiceDesc := &RestServiceDesc{
			HandlerType: (*TestService)(nil),
			Methods:     []RestMethodDesc{},
		}
		server.RegisterRestService(restServiceDesc, nil)
	}
	assert.Equal(t, 3, len(server.restServices))

	// Test multiple raw handler registrations
	var rawHandlers []*RestRawHandlerDesc
	for i := 0; i < 4; i++ {
		rawHandlers = append(rawHandlers, &RestRawHandlerDesc{
			Method:  "GET",
			Path:    fmt.Sprintf("/endpoint/%d", i),
			Handler: nil,
		})
	}
	server.RegisterRestRawHandlers(rawHandlers...)
	assert.Equal(t, 4, len(server.rawHandlers))

	// Test multiple endpoints
	var endpoints []Endpoint
	for i := 0; i < 2; i++ {
		endpoints = append(endpoints, NewMockEndpoint(
			"grpc",
			fmt.Sprintf("localhost:%d", 8080+i),
			pkg.ServerKindRpc,
		))
	}
	server.endpoints = endpoints
	assert.Equal(t, 2, len(server.Endpoints()))
}

// Test interface nil values
func TestInterfaceNilValues(t *testing.T) {
	// Test serverInfo with nil metadata
	si := &serverInfo{
		scheme:   "grpc",
		address:  "localhost:8080",
		svrKind:  pkg.ServerKindRpc,
		metadata: nil,
	}

	metadata := si.Metadata()
	assert.Nil(t, metadata)

	// Test serverInfo with empty metadata
	si.metadata = map[string]string{}
	metadata = si.Metadata()
	assert.NotNil(t, metadata)
	assert.Equal(t, 0, len(metadata))
}
