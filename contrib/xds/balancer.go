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
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// LoadBalancingPolicy 负载均衡策略类型
type LoadBalancingPolicy string

const (
	// PolicyRoundRobin 轮询策略
	PolicyRoundRobin LoadBalancingPolicy = "ROUND_ROBIN"
	// PolicyRandom 随机策略
	PolicyRandom LoadBalancingPolicy = "RANDOM"
	// PolicyLeastRequest 最少请求策略
	PolicyLeastRequest LoadBalancingPolicy = "LEAST_REQUEST"
	// PolicyRingHash 环形哈希策略
	PolicyRingHash LoadBalancingPolicy = "RING_HASH"
	// PolicyMaglev Maglev哈希策略
	PolicyMaglev LoadBalancingPolicy = "MAGLEV"
)

// ClusterConfig 集群配置
type ClusterConfig struct {
	Name                string
	LoadBalancingPolicy LoadBalancingPolicy
	ConnectTimeout      time.Duration
	MaxRequests         uint32
	HealthCheckConfig   *HealthCheckConfig
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Interval           time.Duration
	Timeout            time.Duration
	HealthyThreshold   uint32
	UnhealthyThreshold uint32
}

// Balancer 基于XDS的负载均衡器
type Balancer struct {
	client      *xdsClient
	serviceName string
	cache       SnapshotCache

	mu        sync.RWMutex
	endpoints []resolver.Endpoint
	picker    balancer.Picker
	cluster   *ClusterConfig

	closed int32
}

// NewBalancer 创建XDS负载均衡器
func NewBalancer(client *xdsClient, serviceName string) *Balancer {
	return &Balancer{
		client:      client,
		serviceName: serviceName,
		cache:       client.cache,
		cluster: &ClusterConfig{
			Name:                serviceName,
			LoadBalancingPolicy: PolicyRoundRobin, // 默认轮询
			ConnectTimeout:      5 * time.Second,
			MaxRequests:         1000,
		},
	}
}

// Name 返回负载均衡器名称
func (b *Balancer) Name() string {
	return name
}

// GetPicker 获取选择器
func (b *Balancer) GetPicker() balancer.Picker {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.picker
}

// Update 更新配置
func (b *Balancer) Update(configValues config.Values) {
	b.refreshEndpoints()
}

// Close 关闭负载均衡器
func (b *Balancer) Close() error {
	if atomic.CompareAndSwapInt32(&b.closed, 0, 1) {
		logger.InfoField("closing XDS balancer",
			logger.String("service", b.serviceName))
	}
	return nil
}

// refreshEndpoints 刷新端点列表
func (b *Balancer) refreshEndpoints() {
	// 从ResourceManager获取集群配置
	clusterManager := b.client.GetResourceManager(resource.ClusterType)
	if clusterManager == nil {
		logger.DebugField("cluster manager not found for balancer",
			logger.String("service", b.serviceName))
		return
	}

	// 获取端点信息
	endpointManager := b.client.GetResourceManager(resource.EndpointType)
	if endpointManager == nil {
		logger.DebugField("endpoint manager not found for balancer",
			logger.String("service", b.serviceName))
		return
	}

	endpoints := endpointManager.GetResources()

	b.mu.Lock()
	defer b.mu.Unlock()

	// 转换为框架端点格式
	var newEndpoints []resolver.Endpoint
	for _, resource := range endpoints {
		convertedEndpoints := b.convertResourceToEndpoints(resource)
		newEndpoints = append(newEndpoints, convertedEndpoints...)
	}

	b.endpoints = newEndpoints

	// 根据集群配置创建相应的选择器
	b.picker = b.createPicker()

	logger.InfoField("balancer updated",
		logger.String("service", b.serviceName),
		logger.Int("endpoints", len(newEndpoints)),
		logger.String("lb_policy", string(b.cluster.LoadBalancingPolicy)))
}

// convertResourceToEndpoints 转换资源为端点
func (b *Balancer) convertResourceToEndpoints(resource types.Resource) []resolver.Endpoint {
	// 简化实现：基于资源类型创建端点
	// 实际生产环境应该解析具体的XDS资源格式

	endpoints := make([]resolver.Endpoint, 0)

	// 模拟端点创建
	endpoint := &XDSEndpoint{
		address:  fmt.Sprintf("%s-endpoint-%d", b.serviceName, rand.Intn(10)),
		protocol: "grpc",
		metadata: map[string]interface{}{
			"cluster": b.cluster.Name,
			"policy":  string(b.cluster.LoadBalancingPolicy),
			"weight":  100,
			"health":  "HEALTHY",
			"region":  "us-west-2",
		},
	}

	endpoints = append(endpoints, endpoint)
	return endpoints
}

// createPicker 创建选择器
func (b *Balancer) createPicker() balancer.Picker {
	switch b.cluster.LoadBalancingPolicy {
	case PolicyRoundRobin:
		return NewRoundRobinPicker(b.endpoints)
	case PolicyRandom:
		return NewRandomPicker(b.endpoints)
	case PolicyLeastRequest:
		return NewLeastRequestPicker(b.endpoints)
	case PolicyRingHash:
		return NewRingHashPicker(b.endpoints)
	case PolicyMaglev:
		return NewMaglevPicker(b.endpoints)
	default:
		// 默认使用轮询
		return NewRoundRobinPicker(b.endpoints)
	}
}

// SetLoadBalancingPolicy 设置负载均衡策略
func (b *Balancer) SetLoadBalancingPolicy(policy LoadBalancingPolicy) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cluster.LoadBalancingPolicy = policy
	b.picker = b.createPicker()

	logger.InfoField("load balancing policy updated",
		logger.String("service", b.serviceName),
		logger.String("policy", string(policy)))
}

// SetClusterConfig 设置集群配置
func (b *Balancer) SetClusterConfig(config *ClusterConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cluster = config
	b.picker = b.createPicker()

	logger.InfoField("cluster config updated",
		logger.String("service", b.serviceName),
		logger.String("cluster", config.Name),
		logger.String("policy", string(config.LoadBalancingPolicy)))
}

// GetClusterConfig 获取集群配置
func (b *Balancer) GetClusterConfig() *ClusterConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cluster
}

// PickResult 选择结果
type PickResult struct {
	endpoint resolver.Endpoint
	reporter func(error)
}

// Endpoint 返回端点
func (r *PickResult) Endpoint() resolver.Endpoint {
	return r.endpoint
}

// Report 报告调用结果
func (r *PickResult) Report(err error) {
	if r.reporter != nil {
		r.reporter(err)
	}
}

// RoundRobinPicker 轮询选择器
type RoundRobinPicker struct {
	endpoints []resolver.Endpoint
	index     uint64
}

func NewRoundRobinPicker(endpoints []resolver.Endpoint) *RoundRobinPicker {
	return &RoundRobinPicker{
		endpoints: endpoints,
	}
}

func (p *RoundRobinPicker) Next(ri balancer.RpcInfo) (balancer.PickResult, error) {
	if len(p.endpoints) == 0 {
		return nil, balancer.ErrNoAvailableInstance
	}

	index := atomic.AddUint64(&p.index, 1) - 1
	endpoint := p.endpoints[index%uint64(len(p.endpoints))]

	return &PickResult{
		endpoint: endpoint,
		reporter: func(err error) {
			// 可以实现调用结果统计
			logger.DebugField("round robin call completed",
				logger.String("endpoint", endpoint.GetAddress()),
				logger.Err(err))
		},
	}, nil
}

// RandomPicker 随机选择器
type RandomPicker struct {
	endpoints []resolver.Endpoint
}

func NewRandomPicker(endpoints []resolver.Endpoint) *RandomPicker {
	return &RandomPicker{
		endpoints: endpoints,
	}
}

func (p *RandomPicker) Next(ri balancer.RpcInfo) (balancer.PickResult, error) {
	if len(p.endpoints) == 0 {
		return nil, balancer.ErrNoAvailableInstance
	}

	index := rand.Intn(len(p.endpoints))
	endpoint := p.endpoints[index]

	return &PickResult{
		endpoint: endpoint,
		reporter: func(err error) {
			logger.DebugField("random call completed",
				logger.String("endpoint", endpoint.GetAddress()),
				logger.Err(err))
		},
	}, nil
}

// LeastRequestPicker 最少请求选择器
type LeastRequestPicker struct {
	endpoints []resolver.Endpoint
	requests  []uint64
}

func NewLeastRequestPicker(endpoints []resolver.Endpoint) *LeastRequestPicker {
	return &LeastRequestPicker{
		endpoints: endpoints,
		requests:  make([]uint64, len(endpoints)),
	}
}

func (p *LeastRequestPicker) Next(ri balancer.RpcInfo) (balancer.PickResult, error) {
	if len(p.endpoints) == 0 {
		return nil, balancer.ErrNoAvailableInstance
	}

	// 找到请求数最少的端点
	minIndex := 0
	minRequests := atomic.LoadUint64(&p.requests[0])

	for i := 1; i < len(p.endpoints); i++ {
		reqs := atomic.LoadUint64(&p.requests[i])
		if reqs < minRequests {
			minRequests = reqs
			minIndex = i
		}
	}

	// 增加请求计数
	atomic.AddUint64(&p.requests[minIndex], 1)

	endpoint := p.endpoints[minIndex]
	return &PickResult{
		endpoint: endpoint,
		reporter: func(err error) {
			// 减少请求计数
			atomic.AddUint64(&p.requests[minIndex], ^uint64(0))
			logger.DebugField("least request call completed",
				logger.String("endpoint", endpoint.GetAddress()),
				logger.Uint64("active_requests", atomic.LoadUint64(&p.requests[minIndex])),
				logger.Err(err))
		},
	}, nil
}

// MaglevPicker Maglev一致性哈希选择器
type MaglevPicker struct {
	endpoints []resolver.Endpoint
	hash      *MaglevHash
}

func NewMaglevPicker(endpoints []resolver.Endpoint) *MaglevPicker {
	return &MaglevPicker{
		endpoints: endpoints,
		hash:      NewMaglevHash(len(endpoints)),
	}
}

func (p *MaglevPicker) Next(ri balancer.RpcInfo) (balancer.PickResult, error) {
	if len(p.endpoints) == 0 {
		return nil, balancer.ErrNoAvailableInstance
	}

	// 基于请求信息计算哈希
	hashKey := fmt.Sprintf("%s:%v", ri.Method, ri.Ctx.Value("request_key"))
	hash := p.hash.Lookup(hashKey)

	endpoint := p.endpoints[hash%uint64(len(p.endpoints))]
	return &PickResult{
		endpoint: endpoint,
		reporter: func(err error) {
			logger.DebugField("maglev call completed",
				logger.String("endpoint", endpoint.GetAddress()),
				logger.String("hash_key", hashKey),
				logger.Uint64("hash_result", hash),
				logger.Err(err))
		},
	}, nil
}

// RingHashPicker 环形哈希选择器
type RingHashPicker struct {
	endpoints []resolver.Endpoint
	ring      *RingHash
}

func NewRingHashPicker(endpoints []resolver.Endpoint) *RingHashPicker {
	return &RingHashPicker{
		endpoints: endpoints,
		ring:      NewRingHash(endpoints),
	}
}

func (p *RingHashPicker) Next(ri balancer.RpcInfo) (balancer.PickResult, error) {
	if len(p.endpoints) == 0 {
		return nil, balancer.ErrNoAvailableInstance
	}

	// 基于请求信息计算哈希
	hashKey := fmt.Sprintf("%s:%v", ri.Method, ri.Ctx.Value("request_key"))
	endpoint := p.ring.Lookup(hashKey)

	return &PickResult{
		endpoint: endpoint,
		reporter: func(err error) {
			logger.DebugField("ring hash call completed",
				logger.String("endpoint", endpoint.GetAddress()),
				logger.String("hash_key", hashKey),
				logger.Err(err))
		},
	}, nil
}
