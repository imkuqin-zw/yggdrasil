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
	"net"
	"os"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

// Node 表示XDS节点信息
type Node struct {
	ID       string
	Cluster  string
	Metadata map[string]string
	Locality *Locality
}

// Locality 位置信息
type Locality struct {
	Region  string
	Zone    string
	SubZone string
}

// NewNode 创建新的节点
func NewNode(id, cluster string, metadata map[string]string) *Node {
	node := &Node{
		ID:       id,
		Cluster:  cluster,
		Metadata: metadata,
	}

	// 自动获取本地信息
	if hostname, err := os.Hostname(); err == nil {
		if node.Metadata == nil {
			node.Metadata = make(map[string]string)
		}
		node.Metadata["hostname"] = hostname
	}

	// 尝试获取网络位置信息
	if locality := detectLocality(); locality != nil {
		node.Locality = locality
	}

	return node
}

// ToEnvoyNode 转换为Envoy节点格式（简化实现）
func (n *Node) ToEnvoyNode() *Node {
	// 简化实现：直接返回自身，不依赖Envoy特定的protobuf类型
	return n
}

// detectLocality 检测本地位置信息
func detectLocality() *Locality {
	// 这里可以实现基于云环境、Kubernetes等的位置检测
	// 简化实现，实际可以通过环境变量、元数据服务等获取

	region := os.Getenv("REGION")
	zone := os.Getenv("ZONE")
	subzone := os.Getenv("SUBZONE")

	if region == "" && zone == "" && subzone == "" {
		// 尝试从网络信息推断
		if addrs, err := net.InterfaceAddrs(); err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						// 简单的内网IP地址位置推断
						if isPrivateIP(ipNet.IP) {
							region = "us-west-2"
							zone = "us-west-2a"
							subzone = "us-west-2a-rack1"
						}
						break
					}
				}
			}
		}
	}

	if region != "" || zone != "" || subzone != "" {
		locality := &Locality{
			Region:  region,
			Zone:    zone,
			SubZone: subzone,
		}

		logger.DebugField("detected node locality",
			logger.String("region", region),
			logger.String("zone", zone),
			logger.String("subzone", subzone))

		return locality
	}

	return nil
}

// isPrivateIP 检查是否为私有IP地址
func isPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	return false
}

// UpdateMetadata 更新节点元数据
func (n *Node) UpdateMetadata(key, value string) {
	if n.Metadata == nil {
		n.Metadata = make(map[string]string)
	}
	n.Metadata[key] = value
}

// GetMetadata 获取节点元数据
func (n *Node) GetMetadata(key string) (string, bool) {
	if n.Metadata == nil {
		return "", false
	}
	value, exists := n.Metadata[key]
	return value, exists
}

// ToNodeInfo 转换为节点信息结构（用于框架内部使用）
func (n *Node) ToNodeInfo() NodeInfo {
	return NodeInfo{
		ID:       n.ID,
		Cluster:  n.Cluster,
		Metadata: n.Metadata,
		Locality: n.Locality,
	}
}

// NodeInfo 节点信息结构
type NodeInfo struct {
	ID       string
	Cluster  string
	Metadata map[string]string
	Locality *Locality
}

// Copy 复制节点信息
func (ni NodeInfo) Copy() NodeInfo {
	metadataCopy := make(map[string]string)
	for k, v := range ni.Metadata {
		metadataCopy[k] = v
	}

	var localityCopy *Locality
	if ni.Locality != nil {
		localityCopy = &Locality{
			Region:  ni.Locality.Region,
			Zone:    ni.Locality.Zone,
			SubZone: ni.Locality.SubZone,
		}
	}

	return NodeInfo{
		ID:       ni.ID,
		Cluster:  ni.Cluster,
		Metadata: metadataCopy,
		Locality: localityCopy,
	}
}

// Merge 合并节点信息
func (ni NodeInfo) Merge(other NodeInfo) NodeInfo {
	result := ni.Copy()

	if other.ID != "" {
		result.ID = other.ID
	}
	if other.Cluster != "" {
		result.Cluster = other.Cluster
	}

	for k, v := range other.Metadata {
		result.Metadata[k] = v
	}

	if other.Locality != nil {
		if result.Locality == nil {
			result.Locality = &Locality{}
		}
		if other.Locality.Region != "" {
			result.Locality.Region = other.Locality.Region
		}
		if other.Locality.Zone != "" {
			result.Locality.Zone = other.Locality.Zone
		}
		if other.Locality.SubZone != "" {
			result.Locality.SubZone = other.Locality.SubZone
		}
	}

	return result
}

// String 返回节点信息的字符串表示
func (ni NodeInfo) String() string {
	return fmt.Sprintf("Node{id=%s, cluster=%s, region=%s, zone=%s}",
		ni.ID, ni.Cluster, ni.GetRegion(), ni.GetZone())
}

// GetRegion 获取区域信息
func (ni NodeInfo) GetRegion() string {
	if ni.Locality != nil {
		return ni.Locality.Region
	}
	return ""
}

// GetZone 获取区域信息
func (ni NodeInfo) GetZone() string {
	if ni.Locality != nil {
		return ni.Locality.Zone
	}
	return ""
}

// GetSubZone 获取子区域信息
func (ni NodeInfo) GetSubZone() string {
	if ni.Locality != nil {
		return ni.Locality.SubZone
	}
	return ""
}

// HasMetadata 检查是否包含指定元数据
func (ni NodeInfo) HasMetadata(key string) bool {
	_, exists := ni.Metadata[key]
	return exists
}

// GetMetadataValue 获取元数据值
func (ni NodeInfo) GetMetadataValue(key string) string {
	return ni.Metadata[key]
}

// SetMetadataValue 设置元数据值
func (ni *NodeInfo) SetMetadataValue(key, value string) {
	if ni.Metadata == nil {
		ni.Metadata = make(map[string]string)
	}
	ni.Metadata[key] = value
}
