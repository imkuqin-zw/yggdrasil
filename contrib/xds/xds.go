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

package xds

import (
	"context"
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
)

var (
	// 全局XDS客户端实例
	globalClient Client
)

// init 初始化XDS组件
func init() {
	// 注册Resolver构建器
	resolver.RegisterBuilder(name, newResolver)

	// 注册Balancer构建器
	balancer.RegisterBuilder(name, newBalancer)

	// 注册优雅关闭
	defers.Register(func() error {
		if globalClient != nil {
			return globalClient.Stop()
		}
		return nil
	})
}

// newResolver 创建Resolver构建器
func newResolver(_ string) (resolver.Resolver, error) {
	if globalClient == nil {
		return nil, fmt.Errorf("XDS client not initialized")
	}
	return globalClient.GetResolver(), nil
}

// newBalancer 创建Balancer构建器
func newBalancer(serviceName string) balancer.Balancer {
	if globalClient == nil {
		logger.WarnField("XDS client not initialized, using default balancer")
		return &balancer.RoundRobin{}
	}
	return globalClient.GetBalancer(serviceName)
}

// Initialize 初始化XDS客户端
func Initialize(ctx context.Context) error {
	logger.InfoField("initializing XDS client")

	client, err := NewClient()
	if err != nil {
		return fmt.Errorf("failed to create XDS client: %w", err)
	}

	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start XDS client: %w", err)
	}

	globalClient = client
	logger.InfoField("XDS client initialized successfully")
	return nil
}

// GetClient 获取全局XDS客户端
func GetClient() Client {
	return globalClient
}

// IsInitialized 检查是否已初始化
func IsInitialized() bool {
	return globalClient != nil && globalClient.IsReady()
}

// Shutdown 关闭XDS客户端
func Shutdown() error {
	if globalClient != nil {
		return globalClient.Stop()
	}
	return nil
}
