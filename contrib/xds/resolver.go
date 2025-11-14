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
	"sync"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	resolver2 "github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// Resolver 基于XDS的服务发现解析器
type Resolver struct {
	client *xdsClient
	cache  SnapshotCache

	mu       sync.RWMutex
	watchers map[string]*ServiceWatcher
	closed   bool
}

// ServiceWatcher 服务监听器
type ServiceWatcher struct {
	resolver    *Resolver
	serviceName string
	cancel      context.CancelFunc
	endpoints   []resolver2.Endpoint
	callbacks   []func([]resolver2.Endpoint)
}

// NewResolver 创建XDS解析器
func NewResolver(client *xdsClient) *Resolver {
	return &Resolver{
		client:   client,
		cache:    client.cache,
		watchers: make(map[string]*ServiceWatcher),
	}
}

// Name 返回解析器名称
func (r *Resolver) Name() string {
	return name
}

// AddWatch 添加服务监听
func (r *Resolver) AddWatch(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("resolver is closed")
	}

	// 检查是否已在监听
	if _, exists := r.watchers[serviceName]; exists {
		return nil
	}

	logger.InfoField("adding service watch", logger.String("service", serviceName))

	// 创建服务监听器
	ctx, cancel := context.WithCancel(context.Background())
	watcher := &ServiceWatcher{
		resolver:    r,
		serviceName: serviceName,
		cancel:      cancel,
		callbacks:   make([]func([]resolver2.Endpoint), 0),
	}

	// 启动监听
	if err := watcher.start(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to start service watcher: %w", err)
	}

	r.watchers[serviceName] = watcher
	return nil
}

// DelWatch 删除服务监听
func (r *Resolver) DelWatch(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	watcher, exists := r.watchers[serviceName]
	if !exists {
		return nil
	}

	logger.InfoField("removing service watch", logger.String("service", serviceName))

	// 停止监听
	watcher.stop()
	delete(r.watchers, serviceName)

	return nil
}

// Close 关闭解析器
func (r *Resolver) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	logger.InfoField("closing XDS resolver")

	r.closed = true

	// 停止所有监听器
	for serviceName, watcher := range r.watchers {
		watcher.stop()
		delete(r.watchers, serviceName)
	}

	return nil
}

// start 启动服务监听
func (w *ServiceWatcher) start(ctx context.Context) error {
	// 启动定期刷新goroutine
	xgo.Go(func() {
		w.refreshLoop(ctx)
	}, nil)

	return nil
}

// stop 停止服务监听
func (w *ServiceWatcher) stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

// refreshLoop 定期刷新服务端点
func (w *ServiceWatcher) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒刷新一次
	defer ticker.Stop()

	// 立即执行一次刷新
	w.refreshEndpoints()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.refreshEndpoints()
		}
	}
}

// refreshEndpoints 刷新端点列表
func (w *ServiceWatcher) refreshEndpoints() {
	// 从ResourceManager获取端点资源
	endpointManager := w.resolver.client.GetResourceManager(resource.EndpointType)
	if endpointManager == nil {
		logger.DebugField("endpoint manager not found",
			logger.String("service", w.serviceName))
		return
	}

	resources := endpointManager.GetResources()
	if len(resources) == 0 {
		logger.DebugField("no endpoint resources found",
			logger.String("service", w.serviceName))
		return
	}

	// 转换为框架端点格式
	var newEndpoints []resolver2.Endpoint
	for _, resource := range resources {
		// 简化实现：基于资源名称匹配服务
		resourceName := getResourceName(resource)
		if w.matchesService(resourceName) {
			convertedEndpoints := w.convertResourceToEndpoints(resource)
			newEndpoints = append(newEndpoints, convertedEndpoints...)
		}
	}

	// 检查是否有变化
	if !w.endpointsChanged(newEndpoints) {
		return
	}

	w.endpoints = newEndpoints
	logger.InfoField("endpoints updated",
		logger.String("service", w.serviceName),
		logger.Int("count", len(newEndpoints)))

	// 通知回调函数
	for _, callback := range w.callbacks {
		callback(newEndpoints)
	}
}

// matchesService 检查资源是否匹配服务
func (w *ServiceWatcher) matchesService(resourceName string) bool {
	// 简化匹配逻辑：检查资源名是否包含服务名
	return resourceName == w.serviceName ||
		resourceName == fmt.Sprintf("cluster.%s", w.serviceName) ||
		resourceName == fmt.Sprintf("%s-endpoints", w.serviceName)
}

// convertResourceToEndpoints 转换资源为端点
func (w *ServiceWatcher) convertResourceToEndpoints(resource types.Resource) []resolver2.Endpoint {
	// 简化实现：基于资源类型创建端点
	// 实际生产环境应该解析具体的XDS资源格式

	endpoints := make([]resolver2.Endpoint, 0)

	// 模拟创建端点
	for i := 0; i < 3; i++ {
		endpoint := &XDSEndpoint{
			address:  fmt.Sprintf("%s-%d.example.com:8080", w.serviceName, i),
			protocol: "grpc",
			metadata: map[string]interface{}{
				"service": w.serviceName,
				"weight":  100,
				"region":  "us-west-2",
				"zone":    "us-west-2a",
				"health":  "HEALTHY",
				"version": "v1.0.0",
			},
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

// endpointsChanged 检查端点是否发生变化
func (w *ServiceWatcher) endpointsChanged(newEndpoints []resolver2.Endpoint) bool {
	if len(w.endpoints) != len(newEndpoints) {
		return true
	}

	// 简单比较地址
	oldAddresses := make(map[string]bool)
	for _, endpoint := range w.endpoints {
		oldAddresses[endpoint.GetAddress()] = true
	}

	for _, endpoint := range newEndpoints {
		if !oldAddresses[endpoint.GetAddress()] {
			return true
		}
	}

	return false
}

// AddCallback 添加端点变化回调
func (w *ServiceWatcher) AddCallback(callback func([]resolver2.Endpoint)) {
	w.callbacks = append(w.callbacks, callback)
	// 立即调用一次回调
	callback(w.endpoints)
}

// XDSEndpoint XDS端点实现
type XDSEndpoint struct {
	address  string
	protocol string
	metadata map[string]interface{}
}

// GetAddress 获取地址
func (e *XDSEndpoint) GetAddress() string {
	return e.address
}

// GetProtocol 获取协议
func (e *XDSEndpoint) GetProtocol() string {
	return e.protocol
}

// GetMetadata 获取元数据
func (e *XDSEndpoint) GetMetadata() map[string]interface{} {
	return e.metadata
}

// GetServiceEndpoints 获取服务端点（公共接口）
func (r *Resolver) GetServiceEndpoints(serviceName string) []resolver2.Endpoint {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if watcher, exists := r.watchers[serviceName]; exists {
		return watcher.endpoints
	}

	return nil
}

// RegisterEndpointCallback 注册端点变化回调
func (r *Resolver) RegisterEndpointCallback(serviceName string, callback func([]resolver2.Endpoint)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	watcher, exists := r.watchers[serviceName]
	if !exists {
		return fmt.Errorf("service not watched: %s", serviceName)
	}

	watcher.AddCallback(callback)
	return nil
}

// WatchService 监听服务（用于外部调用）
func (r *Resolver) WatchService(serviceName string, callback func([]resolver2.Endpoint)) error {
	// 如果没有监听器，先创建
	if err := r.AddWatch(serviceName); err != nil {
		return err
	}

	// 注册回调
	return r.RegisterEndpointCallback(serviceName, callback)
}

// UnwatchService 取消监听服务
func (r *Resolver) UnwatchService(serviceName string) error {
	return r.DelWatch(serviceName)
}

// GetAllWatchedServices 获取所有监听的服务
func (r *Resolver) GetAllWatchedServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]string, 0, len(r.watchers))
	for serviceName := range r.watchers {
		services = append(services, serviceName)
	}

	return services
}

// GetStats 获取解析器统计信息
func (r *Resolver) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["watched_services_count"] = len(r.watchers)
	stats["closed"] = r.closed

	serviceStats := make(map[string]interface{})
	for serviceName, watcher := range r.watchers {
		serviceStats[serviceName] = map[string]interface{}{
			"endpoints_count": len(watcher.endpoints),
			"callbacks_count": len(watcher.callbacks),
		}
	}
	stats["services"] = serviceStats

	return stats
}
