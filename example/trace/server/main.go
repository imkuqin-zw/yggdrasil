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

package main

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"

	_ "github.com/imkuqin-zw/yggdrasil/contrib/trace/jaeger"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/grpc/trace"
)

type GreeterImpl struct {
	helloword.UnimplementedGreeterServer
}

func (g GreeterImpl) SayHello(ctx context.Context, request *helloword.HelloRequest) (*helloword.HelloReply, error) {
	return &helloword.HelloReply{Message: request.Name}, nil
}

func (g GreeterImpl) SayHelloStream(server helloword.GreeterSayHelloStreamServer) error {
	panic("implement me")
}

func (g GreeterImpl) SayHelloClientStream(server helloword.GreeterSayHelloClientStreamServer) error {
	panic("implement me")
}

func (g GreeterImpl) SayHelloServerStream(request *helloword.HelloRequest, server helloword.GreeterSayHelloServerStreamServer) error {
	panic("implement me")
}

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		log.Fatal(err)
	}
	grpc.RegisterService(&grpcimpl.GreeterServiceDesc, GreeterImpl{})
	if err := yggdrasil.Run("sample"); err != nil {
		log.Fatal(err)
	}
}
