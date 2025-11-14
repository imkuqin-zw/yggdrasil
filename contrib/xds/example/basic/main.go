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
	"log"
	"time"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/contrib/xds"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

// 定义gRPC服务
type GreeterServer struct {
	yggdrasil.UnimplementedGreeterServer
}

func (s *GreeterServer) SayHello(ctx context.Context, req *yggdrasil.HelloRequest) (*yggdrasil.HelloReply, error) {
	logger.InfoField("received request",
		logger.String("name", req.Name))

	return &yggdrasil.HelloReply{
		Message: "Hello " + req.Name,
	}, nil
}

func main() {
	ctx := context.Background()

	// 设置日志级别
	logger.SetLevel(logger.LvInfo)

	logger.InfoField("starting basic XDS example")

	// 初始化XDS客户端
	logger.InfoField("initializing XDS client")
	if err := xds.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize XDS: %v", err)
	}

	// 等待XDS就绪
	logger.InfoField("waiting for XDS to be ready")
	for i := 0; i < 30; i++ {
		if xds.IsInitialized() {
			logger.InfoField("XDS client is ready")
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !xds.IsInitialized() {
		log.Fatalf("XDS client failed to initialize within timeout")
	}

	// 创建gRPC服务实例
	server := &GreeterServer{}

	// 启动Yggdrasil服务
	logger.InfoField("starting Yggdrasil service")
	if err := yggdrasil.Run("greeter-service",
		yggdrasil.WithServiceDesc(&yggdrasil.GreeterServiceDesc, server),
		yggdrasil.WithRestServiceDesc(&yggdrasil.GreeterServiceRestServiceDesc, server),
	); err != nil {
		log.Fatalf("Failed to run service: %v", err)
	}
}
