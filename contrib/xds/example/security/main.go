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
	"fmt"
	"log"
	"time"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/contrib/xds"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	ym "github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// 定义安全的gRPC服务
type SecureService struct {
	helloword.UnimplementedGreeterServer
}

func (s *SecureService) SecureMethod(ctx context.Context, req *helloword.SecureRequest) (*yggdrasil.SecureReply, error) {
	// 从metadata中获取用户信息
	md, ok := ym.FromInContext(ctx)
	if !ok {
		return nil, status.Error(13, "no metadata found")
	}

	// 获取用户信息
	var userID string
	if users, exists := md["x-user"]; exists && len(users) > 0 {
		userID = users[0]
	}

	logger.InfoField("secure method called",
		logger.String("user_id", userID),
		logger.String("message", req.Message))

	return &yggdrasil.SecureReply{
		Message: fmt.Sprintf("Secure response to user %s: %s", userID, req.Message),
		UserID:  userID,
	}, nil
}

func main() {
	ctx := context.Background()

	// 设置日志级别
	logger.SetLevel(logger.LvInfo)

	logger.InfoField("starting security XDS example")

	// 初始化XDS客户端
	if err := xds.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize XDS: %v", err)
	}

	// 等待XDS就绪
	for i := 0; i < 30; i++ {
		if xds.IsInitialized() {
			logger.InfoField("XDS client is ready")
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !xds.IsInitialized() {
		log.Fatalf("XDS client failed to initialize")
	}

	// 创建安全服务实例
	server := &SecureService{}

	// 启动Yggdrasil服务
	if err := yggdrasil.Run("secure-service",
		yggdrasil.WithServiceDesc(&yggdrasil.SecureServiceDesc, server),
		yggdrasil.WithRestServiceDesc(&yggdrasil.SecureServiceRestServiceDesc, server),
	); err != nil {
		log.Fatalf("Failed to run service: %v", err)
	}
}

// 客户端示例代码
func exampleSecureClient() {
	// 创建TLS凭证
	creds, err := credentials.NewClientTLSFromFile("/etc/certs/ca.crt", "secure-service")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials: %v", err)
	}

	// 创建客户端连接
	conn, err := yggdrasil.NewClient("secure-service",
		yggdrasil.WithTransportCredentials(creds),
		yggdrasil.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			// 添加认证token
			md := metadata.New(map[string]string{
				"authorization": "Bearer my-jwt-token",
				"x-user":        "user123",
			})
			ctx = metadata.NewOutgoingContext(ctx, md)
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := yggdrasil.NewSecureServiceClient(conn)

	// 调用安全方法
	resp, err := client.SecureMethod(ctx, &yggdrasil.SecureRequest{
		Message: "Hello secure service",
	})
	if err != nil {
		log.Fatalf("Secure call failed: %v", err)
	}

	logger.InfoField("secure response",
		logger.String("message", resp.Message),
		logger.String("user_id", resp.UserID))
}
