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
	flag2 "flag"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/contrib/polaris"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/polaris"
	"github.com/imkuqin-zw/yggdrasil/contrib/polaris/exmaple/common/proto"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/flag"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/interceptor/logging"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc"
)

type GreeterCircuitBreakerService struct {
	helloword.UnimplementedGreeterServer
}

func (h *GreeterCircuitBreakerService) SayHello(_ context.Context, request *helloword.HelloRequest) (*helloword.HelloReply, error) {
	return &helloword.HelloReply{Message: request.Name}, nil
	//return nil, status.New(code.Code_INTERNAL, errors.New("error"))
}

var (
	_ = flag2.String("server-name", "0", "server name")
)

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		logger.FatalField("fault to load config file", logger.Err(err))
	}
	if err := config.LoadSource(flag.NewSource()); err != nil {
		logger.FatalField("fault to load config file", logger.Err(err))
	}

	loadConfig()

	name := config.Get("server.name").String("0")
	svrName := "github.com.imkuqin_zw.yggdrasil_polaris.example.server." + name

	if err := yggdrasil.Run(svrName,
		yggdrasil.WithServiceDesc(&helloword.GreeterServiceDesc, &GreeterCircuitBreakerService{}),
	); err != nil {
		logger.FatalField("the application was ended forcefully ", logger.Err(err))
	}
}

func loadConfig() {
	namespace := config.Get(config.KeyAppNamespace).String("default")
	sourceFile := config.GetString("polaris.source", "polaris_demo_server.yaml")
	sc, err := polaris.NewConfig(namespace, "yggdrasil", sourceFile)
	if err != nil {
		logger.FatalField("fault to create polaris data source", logger.Err(err))
	}
	if err := config.LoadSource(sc); err != nil {
		logger.FatalField("fault to load polaris data config", logger.Err(err))
	}
}
