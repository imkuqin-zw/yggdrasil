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
	"strings"
	"sync"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// SnapshotCache XDS快照缓存接口
type SnapshotCache interface {
	// SetSnapshot 设置快照
	SetSnapshot(ctx context.Context, node string, snapshot cache.Snapshot) error

	// GetSnapshot 获取快照
	GetSnapshot(node string) (cache.Snapshot, error)

	// Clear 清除缓存
	Clear(node string)

	// HasInitialResources 检查是否有初始资源
	HasInitialResources() bool

	// GetResources 获取指定类型的资源
	GetResources(typ resource.Type) []types.Resource

	// SetResources 设置指定类型的资源
	SetResources(typ resource.Type, resources []types.Resource) error
}

// SnapshotCacheImpl 快照缓存实现
type SnapshotCacheImpl struct {
	mu        sync.RWMutex
	snapshots map[string]cache.Snapshot
	resources map[resource.Type]map[string]types.Resource
	hasInit   bool
}

// NewSnapshotCache 创建快照缓存
func NewSnapshotCache(ads bool, node *Node) SnapshotCache {
	return &SnapshotCacheImpl{
		snapshots: make(map[string]cache.Snapshot),
		resources: make(map[resource.Type]map[string]types.Resource),
	}
}

// SetSnapshot 设置快照
func (s *SnapshotCacheImpl) SetSnapshot(ctx context.Context, nodeID string, snapshot cache.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证快照
	if err := snapshot.Consistent(); err != nil {
		return fmt.Errorf("snapshot is not consistent: %w", err)
	}

	// 设置快照
	s.snapshots[nodeID] = snapshot

	// 更新资源索引
	s.updateResources(snapshot)

	// 标记已有初始资源
	s.hasInit = true

	logger.DebugField("snapshot set",
		logger.String("node", nodeID),
		logger.Any("version", snapshot.GetVersion(resource.ListenerType)),
		logger.Any("listener_count", len(snapshot.GetResources(resource.ListenerType))),
		logger.Any("route_count", len(snapshot.GetResources(resource.RouteType))),
		logger.Any("cluster_count", len(snapshot.GetResources(resource.ClusterType))),
		logger.Any("endpoint_count", len(snapshot.GetResources(resource.EndpointType))))

	return nil
}

// GetSnapshot 获取快照
func (s *SnapshotCacheImpl) GetSnapshot(nodeID string) (cache.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, exists := s.snapshots[nodeID]
	if !exists {
		return cache.Snapshot{}, fmt.Errorf("snapshot not found for node: %s", nodeID)
	}

	return snapshot, nil
}

// Clear 清除缓存
func (s *SnapshotCacheImpl) Clear(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.snapshots, nodeID)
	s.hasInit = false
}

// HasInitialResources 检查是否有初始资源
func (s *SnapshotCacheImpl) HasInitialResources() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasInit
}

// GetResources 获取指定类型的资源
func (s *SnapshotCacheImpl) GetResources(typ resource.Type) []types.Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resources, exists := s.resources[typ]; exists {
		result := make([]types.Resource, 0, len(resources))
		for _, resource := range resources {
			result = append(result, resource)
		}
		return result
	}

	return nil
}

// SetResources 设置指定类型的资源
func (s *SnapshotCacheImpl) SetResources(typ resource.Type, resources []types.Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.resources[typ] == nil {
		s.resources[typ] = make(map[string]types.Resource)
	}

	// 清空现有资源
	s.resources[typ] = make(map[string]types.Resource)

	// 添加新资源
	for _, resource := range resources {
		if namedResource, ok := resource.(types.ResourceWithName); ok {
			s.resources[typ][namedResource.GetName()] = resource
		}
	}

	s.hasInit = true
	return nil
}

// updateResources 更新资源索引
func (s *SnapshotCacheImpl) updateResources(snapshot cache.Snapshot) {
	// 为每种资源类型更新索引
	resourceTypes := []resource.Type{
		resource.ListenerType,
		resource.RouteType,
		resource.ClusterType,
		resource.EndpointType,
	}

	for _, typ := range resourceTypes {
		if s.resources[typ] == nil {
			s.resources[typ] = make(map[string]types.Resource)
		}

		resources := snapshot.GetResources(typ)
		for _, resource := range resources {
			name := getResourceName(resource)
			s.resources[typ][name] = resource
		}
	}
}

// getResourceName 获取资源名称
func getResourceName(resource types.Resource) string {
	// 根据资源类型提取名称
	switch r := resource.(type) {
	case interface{ GetName() string }:
		return r.GetName()
	default:
		// 使用类型断言或其他方式获取名称
		return fmt.Sprintf("%T", resource)
	}
}

// CreateSnapshot 创建快照
func CreateSnapshot(version string, listeners, routes, clusters, endpoints []types.Resource) (*cache.Snapshot, error) {
	// 创建资源映射
	resources := make(map[resource.Type][]types.Resource)
	resources[resource.ListenerType] = listeners
	resources[resource.RouteType] = routes
	resources[resource.ClusterType] = clusters
	resources[resource.EndpointType] = endpoints

	// 创建快照
	snapshot, err := cache.NewSnapshot(version, resources)
	if err != nil {
		return &cache.Snapshot{}, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return snapshot, nil
}

// GenerateVersion 生成版本号
func GenerateVersion() string {
	return fmt.Sprintf("v%d", time.Now().UnixNano())
}

// FilterResources 过滤资源
func FilterResources(resources []types.Resource, namePattern string) []types.Resource {
	if namePattern == "*" {
		return resources
	}

	var filtered []types.Resource
	pattern := strings.ToLower(namePattern)

	for _, resource := range resources {
		name := strings.ToLower(getResourceName(resource))
		if strings.Contains(name, pattern) {
			filtered = append(filtered, resource)
		}
	}

	return filtered
}
