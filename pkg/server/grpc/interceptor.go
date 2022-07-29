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

package grpc

import (
	"google.golang.org/grpc"
)

var unaryInterceptor = map[string]func() grpc.UnaryServerInterceptor{}
var streamInterceptor = map[string]func() grpc.StreamServerInterceptor{}
var serverOptions map[string]func() grpc.ServerOption

func RegisterUnaryInterceptor(name string, interceptor func() grpc.UnaryServerInterceptor) {
	unaryInterceptor[name] = interceptor
}

func RegisterStreamInterceptor(name string, interceptor func() grpc.StreamServerInterceptor) {
	streamInterceptor[name] = interceptor
}

func RegisterServerOptions(name string, opt func() grpc.ServerOption) {
	serverOptions[name] = opt
}
