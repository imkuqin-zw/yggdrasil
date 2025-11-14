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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// Client XDS客户端接口
type Client interface {
	// Start 启动XDS客户端
	Start(ctx context.Context) error

	// Stop 停止XDS客户端
	Stop() error

	// GetResolver 获取Resolver
	GetResolver() *Resolver

	// GetBalancer 获取Balancer
	GetBalancer(serviceName string) *Balancer

	// GetNode 获取节点信息
	GetNode() *Node

	// IsReady 检查客户端是否就绪
	IsReady() bool
}

// xdsClient XDS客户端实现
type xdsClient struct {
	config *Config
	node   *Node
	cache  SnapshotCache
	conn   *grpc.ClientConn

	// 简化的实现 - 直接管理资源而不依赖复杂的XDS客户端库
	resourceManagers map[resource.Type]*ResourceManager

	// 状态管理
	mu        sync.RWMutex
	ready     bool
	stopped   bool
	resolver  *Resolver
	balancers map[string]*Balancer

	// 回调函数
	onConfigUpdate func(types.Resource)
}

// ResourceManager 资源管理器
type ResourceManager struct {
	typ       resource.Type
	resources map[string]types.Resource
	version   string
	callbacks []func(types.Resource)
	mu        sync.RWMutex
}

// NewResourceManager 创建资源管理器
func NewResourceManager(typ resource.Type) *ResourceManager {
	return &ResourceManager{
		typ:       typ,
		resources: make(map[string]types.Resource),
		callbacks: make([]func(types.Resource), 0),
	}
}

// UpdateResources 更新资源
func (rm *ResourceManager) UpdateResources(version string, resources []types.Resource) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.version = version
	rm.resources = make(map[string]types.Resource)
	for _, resource := range resources {
		name := getResourceName(resource)
		rm.resources[name] = resource

		// 通知回调
		for _, callback := range rm.callbacks {
			callback(resource)
		}
	}
}

// GetResources 获取资源
func (rm *ResourceManager) GetResources() []types.Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make([]types.Resource, 0, len(rm.resources))
	for _, resource := range rm.resources {
		result = append(result, resource)
	}
	return result
}

// AddCallback 添加回调
func (rm *ResourceManager) AddCallback(callback func(types.Resource)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.callbacks = append(rm.callbacks, callback)
}

// NewClient 创建XDS客户端
func NewClient() (Client, error) {
	// 加载配置
	cfg := &Config{}
	if err := config.Get("xds").Scan(cfg); err != nil {
		logger.WarnField("failed to load xds config, using default", logger.Err(err))
		cfg = defaultConfig
	}

	// 创建节点信息
	node := &Node{
		ID:       cfg.NodeInfo.Id,
		Cluster:  cfg.NodeInfo.Cluster,
		Metadata: cfg.NodeInfo.Metadata,
	}

	// 创建快照缓存
	cache := NewSnapshotCache(true, node)

	// 创建资源管理器
	resourceManagers := map[resource.Type]*ResourceManager{
		resource.ListenerType: NewResourceManager(resource.ListenerType),
		resource.RouteType:    NewResourceManager(resource.RouteType),
		resource.ClusterType:  NewResourceManager(resource.ClusterType),
		resource.EndpointType: NewResourceManager(resource.EndpointType),
	}

	client := &xdsClient{
		config:           cfg,
		node:             node,
		cache:            cache,
		resourceManagers: resourceManagers,
		balancers:        make(map[string]*Balancer),
	}

	// 初始化资源监听
	client.initResourceWatchers()

	return client, nil
}

// initResourceWatchers 初始化资源监听器
func (c *xdsClient) initResourceWatchers() {
	for typ, manager := range c.resourceManagers {
		manager.AddCallback(func(resource types.Resource) {
			logger.DebugField("resource updated",
				logger.String("type", string(typ)),
				logger.String("name", getResourceName(resource)))

			// 更新缓存
			c.updateSnapshot()

			// 触发全局回调
			c.onConfigUpdate(resource)
		})
	}
}

// Start 启动XDS客户端
func (c *xdsClient) Start(ctx context.Context) error {
	logger.InfoField("starting XDS client")

	// 建立与管理服务器的连接
	if err := c.connectManagementServers(); err != nil {
		return fmt.Errorf("failed to connect to management servers: %w", err)
	}

	// 启动资源同步
	if err := c.startResourceSync(ctx); err != nil {
		return fmt.Errorf("failed to start resource sync: %w", err)
	}

	// 等待初始配置加载
	if err := c.waitForInitialResources(ctx); err != nil {
		return fmt.Errorf("failed to wait for initial resources: %w", err)
	}

	c.mu.Lock()
	c.ready = true
	c.mu.Unlock()

	// 创建Resolver
	c.resolver = NewResolver(c)

	logger.InfoField("XDS client started successfully")
	return nil
}

// connectManagementServers 连接管理服务器
func (c *xdsClient) connectManagementServers() error {
	// 创建TLS凭证
	tlsConfig := &tls.Config{
		ServerName:         c.config.Security.ServerName,
		InsecureSkipVerify: c.config.Security.InsecureSkipVerify,
	}

	if c.config.Security.TlsEnabled && c.config.Security.CaCertFile != "" {
		caCert, err := ioutil.ReadFile(c.config.Security.CaCertFile)
		if err != nil {
			return fmt.Errorf("failed to read CA cert file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool

		if c.config.Security.CertFile != "" && c.config.Security.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(c.config.Security.CertFile, c.config.Security.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to load client cert: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	// 创建gRPC连接凭证
	var creds credentials.TransportCredentials
	if c.config.Security.TlsEnabled {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	// 创建gRPC连接选项
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试连接第一个可用的管理服务器
	for _, addr := range c.config.ManagementServerAddresses {
		conn, err := grpc.DialContext(ctx, addr, dialOpts...)
		if err != nil {
			logger.WarnField("failed to connect to management server",
				logger.String("address", addr), logger.Err(err))
			continue
		}

		c.conn = conn
		logger.InfoField("connected to management server", logger.String("address", addr))
		return nil
	}

	return fmt.Errorf("failed to connect to any management server")
}

// startResourceSync 启动资源同步
func (c *xdsClient) startResourceSync(ctx context.Context) error {
	// 简化实现：定期从外部配置源同步资源
	// 实际生产环境应该实现完整的XDS协议

	ticker := time.NewTicker(10 * time.Second) // 每10秒同步一次
	defer ticker.Stop()

	// 启动同步goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.syncResources()
			}
		}
	}()

	return nil
}

// syncResources 同步资源
func (c *xdsClient) syncResources() {
	// 这里应该实现真正的XDS协议通信
	// 为了简化，我们创建一些模拟资源

	logger.DebugField("syncing XDS resources")

	// 模拟监听器资源
	if manager, exists := c.resourceManagers[resource.ListenerType]; exists {
		// 这里可以从XDS服务器获取真实的监听器配置
		// 暂时使用空资源
		manager.UpdateResources(fmt.Sprintf("v%d", time.Now().Unix()), []types.Resource{})
	}

	// 模拟路由资源
	if manager, exists := c.resourceManagers[resource.RouteType]; exists {
		manager.UpdateResources(fmt.Sprintf("v%d", time.Now().Unix()), []types.Resource{})
	}

	// 模拟集群资源
	if manager, exists := c.resourceManagers[resource.ClusterType]; exists {
		// 创建模拟集群配置
		mockClusters := c.createMockClusters()
		manager.UpdateResources(fmt.Sprintf("v%d", time.Now().Unix()), mockClusters)
	}

	// 模拟端点资源
	if manager, exists := c.resourceManagers[resource.EndpointType]; exists {
		// 创建模拟端点配置
		mockEndpoints := c.createMockEndpoints()
		manager.UpdateResources(fmt.Sprintf("v%d", time.Now().Unix()), mockEndpoints)
	}
}

// createMockClusters 创建模拟集群
func (c *xdsClient) createMockClusters() []types.Resource {
	// 这里应该返回真实的集群资源
	// 暂时返回空切片
	return []types.Resource{}
}

// createMockEndpoints 创建模拟端点
func (c *xdsClient) createMockEndpoints() []types.Resource {
	// 这里应该返回真实的端点资源
	// 暂时返回空切片
	return []types.Resource{}
}

// updateSnapshot 更新快照
func (c *xdsClient) updateSnapshot() {
	// 收集所有资源
	var listeners, routes, clusters, endpoints []types.Resource

	if manager, exists := c.resourceManagers[resource.ListenerType]; exists {
		listeners = manager.GetResources()
	}
	if manager, exists := c.resourceManagers[resource.RouteType]; exists {
		routes = manager.GetResources()
	}
	if manager, exists := c.resourceManagers[resource.ClusterType]; exists {
		clusters = manager.GetResources()
	}
	if manager, exists := c.resourceManagers[resource.EndpointType]; exists {
		endpoints = manager.GetResources()
	}

	// 创建快照
	version := fmt.Sprintf("v%d", time.Now().Unix())
	snapshot, err := CreateSnapshot(version, listeners, routes, clusters, endpoints)
	if err != nil {
		logger.WarnField("failed to create snapshot", logger.Err(err))
		return
	}

	// 更新缓存
	if err := c.cache.SetSnapshot(context.Background(), c.node.ID, *snapshot); err != nil {
		logger.WarnField("failed to update cache", logger.Err(err))
	}
}

// waitForInitialResources 等待初始资源加载
func (c *xdsClient) waitForInitialResources(ctx context.Context) error {
	timeout := c.config.InitialLoadTimeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for initial resources")
		case <-ticker.C:
			if c.cache.HasInitialResources() {
				return nil
			}
		}
	}
}

// Stop 停止XDS客户端
func (c *xdsClient) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return nil
	}

	logger.InfoField("stopping XDS client")

	c.stopped = true
	c.ready = false

	// 关闭gRPC连接
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			logger.WarnField("failed to close gRPC connection", logger.Err(err))
		}
	}

	logger.InfoField("XDS client stopped")
	return nil
}

// GetResolver 获取Resolver
func (c *xdsClient) GetResolver() *Resolver {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.resolver
}

// GetBalancer 获取Balancer
func (c *xdsClient) GetBalancer(serviceName string) *Balancer {
	c.mu.Lock()
	defer c.mu.Unlock()

	if balancer, exists := c.balancers[serviceName]; exists {
		return balancer
	}

	balancer := NewBalancer(c, serviceName)
	c.balancers[serviceName] = balancer
	return balancer
}

// GetNode 获取节点信息
func (c *xdsClient) GetNode() *Node {
	return c.node
}

// IsReady 检查客户端是否就绪
func (c *xdsClient) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

// OnConfigUpdate 配置更新回调
func (c *xdsClient) OnConfigUpdate(resource types.Resource) {
	if c.onConfigUpdate != nil {
		c.onConfigUpdate(resource)
	}
}

// SetConfigUpdateCallback 设置配置更新回调
func (c *xdsClient) SetConfigUpdateCallback(callback func(types.Resource)) {
	c.onConfigUpdate = callback
}

// GetResourceManager 获取资源管理器
func (c *xdsClient) GetResourceManager(typ resource.Type) *ResourceManager {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.resourceManagers[typ]
}
