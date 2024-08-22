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

	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
)

type methodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor interceptor.UnaryServerInterceptor) (interface{}, error)

// MethodDesc represents an RPC service's method specification.
type MethodDesc struct {
	MethodName string
	Handler    methodHandler
}

// ServiceDesc represents an RPC service's specification.
type ServiceDesc struct {
	ServiceName string
	// The pointer to the service interface. Used to check whether the user
	// provided implementation satisfies the interface requirements.
	HandlerType interface{}
	Methods     []MethodDesc
	Streams     []stream.StreamDesc
	Metadata    interface{}
}

type ServiceInfo struct {
	// Contains the implementation for the methods in this service.
	ServiceImpl interface{}
	Methods     map[string]*MethodDesc
	Streams     map[string]*stream.StreamDesc
	Metadata    interface{}
}

type methodInfo struct {
	MethodName    string `json:"methodName"`
	ServerStreams bool   `json:"serverStreams"`
	ClientStreams bool   `json:"clientStreams"`
}

type RestMethodHandler func(w http.ResponseWriter, r *http.Request, srv interface{}, interceptor interceptor.UnaryServerInterceptor) (interface{}, error)

type RestServiceDesc struct {
	HandlerType interface{}
	Methods     []RestMethodDesc
}

type RestMethodDesc struct {
	Method  string
	Path    string
	Handler RestMethodHandler
}

type restRouterInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type RestRawHandlerDesc struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}
