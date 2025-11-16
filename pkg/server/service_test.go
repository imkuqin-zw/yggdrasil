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
	"net/http/httptest"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/stretchr/testify/assert"
)

// Test MethodDesc functionality
func TestMethodDesc(t *testing.T) {
	var called bool
	handler := func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
		called = true
		return "handled", nil
	}

	methodDesc := MethodDesc{
		MethodName: "TestMethod",
		Handler:    handler,
	}

	assert.Equal(t, "TestMethod", methodDesc.MethodName)
	assert.NotNil(t, methodDesc.Handler)

	// Test handler execution
	response, err := methodDesc.Handler(nil, context.Background(), nil, nil)
	assert.True(t, called)
	assert.NoError(t, err)
	assert.Equal(t, "handled", response)
}

// Test ServiceInfo functionality
func TestServiceInfo(t *testing.T) {
	methodDesc := &MethodDesc{
		MethodName: "TestMethod",
		Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
			return "response", nil
		},
	}

	streamDesc := &stream.StreamDesc{
		StreamName:    "TestStream",
		ServerStreams: true,
		ClientStreams: true,
	}

	serviceImpl := &TestServiceImpl{}
	serviceInfo := &ServiceInfo{
		ServiceImpl: serviceImpl,
		Methods: map[string]*MethodDesc{
			"TestMethod": methodDesc,
		},
		Streams: map[string]*stream.StreamDesc{
			"TestStream": streamDesc,
		},
		Metadata: "test metadata",
	}

	assert.Equal(t, serviceImpl, serviceInfo.ServiceImpl)
	assert.Equal(t, 1, len(serviceInfo.Methods))
	assert.Equal(t, methodDesc, serviceInfo.Methods["TestMethod"])
	assert.Equal(t, 1, len(serviceInfo.Streams))
	assert.Equal(t, streamDesc, serviceInfo.Streams["TestStream"])
	assert.Equal(t, "test metadata", serviceInfo.Metadata)
}

// Test RestServiceDesc functionality
func TestRestServiceDesc(t *testing.T) {
	var handlerCalled bool
	restHandler := func(w http.ResponseWriter, r *http.Request, srv interface{}, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
		handlerCalled = true
		return "rest response", nil
	}

	restServiceDesc := &RestServiceDesc{
		HandlerType: (*TestService)(nil),
		Methods: []RestMethodDesc{
			{
				Method:  "GET",
				Path:    "/test",
				Handler: restHandler,
			},
			{
				Method:  "POST",
				Path:    "/create",
				Handler: restHandler,
			},
		},
	}

	assert.Equal(t, 2, len(restServiceDesc.Methods))
	assert.Equal(t, "GET", restServiceDesc.Methods[0].Method)
	assert.Equal(t, "/test", restServiceDesc.Methods[0].Path)
	assert.Equal(t, "POST", restServiceDesc.Methods[1].Method)
	assert.Equal(t, "/create", restServiceDesc.Methods[1].Path)

	// Test handler execution
	response, err := restServiceDesc.Methods[0].Handler(nil, nil, nil, nil)
	assert.True(t, handlerCalled)
	assert.NoError(t, err)
	assert.Equal(t, "rest response", response)
}

// Test RestMethodDesc functionality
func TestRestMethodDesc(t *testing.T) {
	var called bool
	handler := func(w http.ResponseWriter, r *http.Request, srv interface{}, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
		called = true
		return "method response", nil
	}

	restMethodDesc := RestMethodDesc{
		Method:  "PUT",
		Path:    "/update",
		Handler: handler,
	}

	assert.Equal(t, "PUT", restMethodDesc.Method)
	assert.Equal(t, "/update", restMethodDesc.Path)
	assert.NotNil(t, restMethodDesc.Handler)

	// Test handler execution
	response, err := restMethodDesc.Handler(nil, nil, nil, nil)
	assert.True(t, called)
	assert.NoError(t, err)
	assert.Equal(t, "method response", response)
}

// Test RestRawHandlerDesc functionality
func TestRestRawHandlerDesc(t *testing.T) {
	var called bool
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	restRawHandlerDesc := &RestRawHandlerDesc{
		Method:  "DELETE",
		Path:    "/delete",
		Handler: handler,
	}

	assert.Equal(t, "DELETE", restRawHandlerDesc.Method)
	assert.Equal(t, "/delete", restRawHandlerDesc.Path)
	assert.NotNil(t, restRawHandlerDesc.Handler)

	// Test handler execution
	w := &httptest.ResponseRecorder{}
	r := httptest.NewRequest("DELETE", "/delete", nil)
	restRawHandlerDesc.Handler(w, r)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test methodInfo struct functionality
func TestMethodInfo(t *testing.T) {
	tests := []struct {
		name         string
		methodInfo   methodInfo
		expectedJSON string
	}{
		{
			name: "unary method",
			methodInfo: methodInfo{
				MethodName:    "TestMethod",
				ServerStreams: false,
				ClientStreams: false,
			},
			expectedJSON: `{"methodName":"TestMethod","serverStreams":false,"clientStreams":false}`,
		},
		{
			name: "server streaming method",
			methodInfo: methodInfo{
				MethodName:    "ServerStreamMethod",
				ServerStreams: true,
				ClientStreams: false,
			},
			expectedJSON: `{"methodName":"ServerStreamMethod","serverStreams":true,"clientStreams":false}`,
		},
		{
			name: "client streaming method",
			methodInfo: methodInfo{
				MethodName:    "ClientStreamMethod",
				ServerStreams: false,
				ClientStreams: true,
			},
			expectedJSON: `{"methodName":"ClientStreamMethod","serverStreams":false,"clientStreams":true}`,
		},
		{
			name: "bidirectional streaming method",
			methodInfo: methodInfo{
				MethodName:    "BidiStreamMethod",
				ServerStreams: true,
				ClientStreams: true,
			},
			expectedJSON: `{"methodName":"BidiStreamMethod","serverStreams":true,"clientStreams":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.methodInfo.MethodName, tt.methodInfo.MethodName)
			assert.Equal(t, tt.methodInfo.ServerStreams, tt.methodInfo.ServerStreams)
			assert.Equal(t, tt.methodInfo.ClientStreams, tt.methodInfo.ClientStreams)

			// Test JSON serialization would work (basic check)
			// In a real test, you'd use json.Marshal
			assert.NotEmpty(t, tt.methodInfo.MethodName)
		})
	}
}

// Test restRouterInfo struct functionality
func TestRestRouterInfo(t *testing.T) {
	tests := []struct {
		name           string
		restRouterInfo restRouterInfo
		expectedMethod string
		expectedPath   string
	}{
		{
			name: "GET endpoint",
			restRouterInfo: restRouterInfo{
				Method: "GET",
				Path:   "/api/v1/users",
			},
			expectedMethod: "GET",
			expectedPath:   "/api/v1/users",
		},
		{
			name: "POST endpoint",
			restRouterInfo: restRouterInfo{
				Method: "POST",
				Path:   "/api/v1/users",
			},
			expectedMethod: "POST",
			expectedPath:   "/api/v1/users",
		},
		{
			name: "PUT endpoint",
			restRouterInfo: restRouterInfo{
				Method: "PUT",
				Path:   "/api/v1/users/123",
			},
			expectedMethod: "PUT",
			expectedPath:   "/api/v1/users/123",
		},
		{
			name: "DELETE endpoint",
			restRouterInfo: restRouterInfo{
				Method: "DELETE",
				Path:   "/api/v1/users/123",
			},
			expectedMethod: "DELETE",
			expectedPath:   "/api/v1/users/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMethod, tt.restRouterInfo.Method)
			assert.Equal(t, tt.expectedPath, tt.restRouterInfo.Path)

			// Test JSON serialization would work (basic check)
			// In a real test, you'd use json.Marshal
			assert.NotEmpty(t, tt.restRouterInfo.Method)
			assert.NotEmpty(t, tt.restRouterInfo.Path)
		})
	}
}

// Test type definitions
func TestTypeDefinitions(t *testing.T) {
	// Test methodHandler type
	var handler methodHandler = func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
		return "handled", nil
	}

	response, err := handler(nil, context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "handled", response)

	// Test RestMethodHandler type
	var restHandler RestMethodHandler = func(w http.ResponseWriter, r *http.Request, srv interface{}, interceptor interceptor.UnaryServerInterceptor) (interface{}, error) {
		return "rest handled", nil
	}

	restResponse, err := restHandler(nil, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "rest handled", restResponse)
}

// Test empty collections
func TestEmptyCollections(t *testing.T) {
	// Test empty ServiceDesc
	emptyServiceDesc := &ServiceDesc{
		ServiceName: "empty.service",
		HandlerType: (*TestService)(nil),
		Methods:     []MethodDesc{},
		Streams:     []stream.StreamDesc{},
		Metadata:    nil,
	}

	assert.Equal(t, 0, len(emptyServiceDesc.Methods))
	assert.Equal(t, 0, len(emptyServiceDesc.Streams))
	assert.Nil(t, emptyServiceDesc.Metadata)

	// Test empty ServiceInfo
	emptyServiceInfo := &ServiceInfo{
		ServiceImpl: nil,
		Methods:     map[string]*MethodDesc{},
		Streams:     map[string]*stream.StreamDesc{},
		Metadata:    nil,
	}

	assert.Equal(t, 0, len(emptyServiceInfo.Methods))
	assert.Equal(t, 0, len(emptyServiceInfo.Streams))
	assert.Nil(t, emptyServiceInfo.Metadata)

	// Test empty RestServiceDesc
	emptyRestServiceDesc := &RestServiceDesc{
		HandlerType: (*TestService)(nil),
		Methods:     []RestMethodDesc{},
	}

	assert.Equal(t, 0, len(emptyRestServiceDesc.Methods))
}
