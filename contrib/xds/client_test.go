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
	"testing"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{}

	// 测试默认配置
	assert.Equal(t, []string{"xds-server:15010"}, cfg.ManagementServerAddresses)
	assert.Equal(t, "yggdrasil-client", cfg.NodeInfo.Id)
	assert.Equal(t, "yggdrasil-cluster", cfg.NodeInfo.Cluster)
	assert.True(t, cfg.XdsResources.Ads)
	assert.True(t, cfg.Security.TlsEnabled)
	assert.Equal(t, 15*time.Second, cfg.InitialLoadTimeout)
}

func TestNodeCreation(t *testing.T) {
	node := NewNode("test-node", "test-cluster", map[string]string{
		"version": "1.0.0",
		"app":     "test-app",
	})

	assert.Equal(t, "test-node", node.ID)
	assert.Equal(t, "test-cluster", node.Cluster)
	assert.Equal(t, "1.0.0", node.Metadata["version"])
	assert.Equal(t, "test-app", node.Metadata["app"])

	// 测试转换为Envoy节点
	envoyNode := node.ToEnvoyNode()
	assert.Equal(t, "test-node", envoyNode.ID)
	assert.Equal(t, "test-cluster", envoyNode.Cluster)
	assert.Equal(t, "1.0.0", node.Metadata["version"])
}

func TestSnapshotCache(t *testing.T) {
	node := NewNode("test", "cluster", nil)
	cache := NewSnapshotCache(true, node)

	// 测试初始状态
	assert.False(t, cache.HasInitialResources())

	// 创建测试快照
	snapshot, err := CreateSnapshot("v1", nil, nil, nil, nil)
	assert.NoError(t, err)

	// 设置快照
	ctx := context.Background()
	err = cache.SetSnapshot(ctx, "test-node", snapshot)
	assert.NoError(t, err)
	assert.True(t, cache.HasInitialResources())

	// 获取快照
	retrieved, err := cache.GetSnapshot("test-node")
	assert.NoError(t, err)
	assert.Equal(t, "v1", retrieved.GetVersion(resource.ListenerType))
}

func TestResolverInterface(t *testing.T) {
	// 测试Resolver接口实现
	r := &Resolver{}
	assert.Equal(t, name, r.Name())
}

func TestBalancerInterface(t *testing.T) {
	// 测试Balancer接口实现
	b := &Balancer{}
	assert.Equal(t, name, b.Name())
}

func TestLoadBalancingPolicies(t *testing.T) {
	// 测试各种负载均衡策略
	endpoints := []resolver.Endpoint{
		&XDSEndpoint{address: "localhost:5001"},
		&XDSEndpoint{address: "localhost:5002"},
		&XDSEndpoint{address: "localhost:5003"},
	}

	// 测试轮询选择器
	roundRobinPicker := NewRoundRobinPicker(endpoints)
	for i := 0; i < 10; i++ {
		result, err := roundRobinPicker.Next(balancer.RpcInfo{Method: "test"})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Endpoint())
	}

	// 测试随机选择器
	randomPicker := NewRandomPicker(endpoints)
	for i := 0; i < 10; i++ {
		result, err := randomPicker.Next(balancer.RpcInfo{Method: "test"})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Endpoint())
	}

	// 测试最少请求选择器
	leastRequestPicker := NewLeastRequestPicker(endpoints)
	for i := 0; i < 10; i++ {
		result, err := leastRequestPicker.Next(balancer.RpcInfo{Method: "test"})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Endpoint())

		// 模拟请求完成
		result.Report(nil)
	}
}

func TestHashAlgorithms(t *testing.T) {
	// 测试Maglev哈希
	maglevHash := NewMaglevHash(3)
	hash1 := maglevHash.Lookup("test-key-1")
	hash2 := maglevHash.Lookup("test-key-2")
	hash3 := maglevHash.Lookup("test-key-1") // 相同的键应该产生相同的哈希

	assert.Equal(t, hash1, hash3)
	assert.NotEqual(t, hash1, hash2)
	assert.Less(t, hash1, uint64(3)) // 应该在后端数量范围内

	// 测试环形哈希
	endpoints := []resolver.Endpoint{
		&XDSEndpoint{address: "server1:8080"},
		&XDSEndpoint{address: "server2:8080"},
		&XDSEndpoint{address: "server3:8080"},
	}
	ringHash := NewRingHash(endpoints)

	endpoint1 := ringHash.Lookup("test-key-1")
	endpoint2 := ringHash.Lookup("test-key-2")
	endpoint3 := ringHash.Lookup("test-key-1") // 相同的键应该映射到相同的端点

	assert.NotNil(t, endpoint1)
	assert.NotNil(t, endpoint2)
	assert.Equal(t, endpoint1, endpoint3)
}

func TestXDSEndpoint(t *testing.T) {
	endpoint := &XDSEndpoint{
		address:  "localhost:50051",
		protocol: "grpc",
		metadata: map[string]interface{}{
			"weight": 100,
			"region": "us-west-2",
		},
	}

	assert.Equal(t, "localhost:50051", endpoint.GetAddress())
	assert.Equal(t, "grpc", endpoint.GetProtocol())
	assert.Equal(t, uint32(100), endpoint.GetMetadata()["weight"])
	assert.Equal(t, "us-west-2", endpoint.GetMetadata()["region"])
}

// Mock对象用于测试
type MockXDSClient struct {
	mock.Mock
}

func (m *MockXDSClient) GetResolver() *Resolver {
	args := m.Called()
	return args.Get(0).(*Resolver)
}

func (m *MockXDSClient) GetBalancer(serviceName string) *Balancer {
	args := m.Called(serviceName)
	return args.Get(0).(*Balancer)
}

func (m *MockXDSClient) GetNode() *Node {
	args := m.Called()
	return args.Get(0).(*Node)
}

func (m *MockXDSClient) IsReady() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestIntegration(t *testing.T) {
	// 集成测试：测试各组件之间的协作
	t.Skip("Integration test - requires actual XDS server")

	// 这里可以编写需要真实XDS服务器的集成测试
	// 1. 启动测试XDS服务器
	// 2. 初始化XDS客户端
	// 3. 验证服务发现和负载均衡
	// 4. 清理资源
}

// Benchmark测试
func BenchmarkRoundRobinPicker(b *testing.B) {
	endpoints := make([]resolver.Endpoint, 100)
	for i := 0; i < 100; i++ {
		endpoints[i] = &XDSEndpoint{address: fmt.Sprintf("server-%d:8080", i)}
	}

	picker := NewRoundRobinPicker(endpoints)
	info := balancer.RpcInfo{Method: "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := picker.Next(info)
		if err != nil {
			b.Fatal(err)
		}
		result.Report(nil)
	}
}

func BenchmarkHashFunction(b *testing.B) {
	hash := NewMaglevHash(10)
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("benchmark-key-%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		hash.Lookup(key)
	}
}
