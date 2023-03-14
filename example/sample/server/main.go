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
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/interceptor/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc"
)

type GreeterImpl struct {
	helloword.UnimplementedGreeterServer
}

func (g GreeterImpl) SayHello(_ context.Context, request *helloword.HelloRequest) (*helloword.HelloReply, error) {
	return &helloword.HelloReply{Message: request.Name}, nil
}

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		logger.FatalFiled("fault to load config file", logger.Err(err))
	}
	yggdrasil.Init("github.com.imkuqin_zw.yggdrasil.example.sample")
	if err := yggdrasil.Serve(yggdrasil.WithServiceDesc(&helloword.GreeterServiceDesc, GreeterImpl{})); err != nil {
		logger.FatalFiled("the application was ended forcefully ", logger.Err(err))
		logger.Fatal(err)
	}
}
