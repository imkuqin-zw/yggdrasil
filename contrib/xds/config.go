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
	"time"
)

const (
	name = "xds"
)

// Config XDS配置
type Config struct {
	// ManagementServerAddresses 控制平面地址列表
	ManagementServerAddresses []string `yaml:"management_server_addresses" json:"management_server_addresses"`

	// NodeInfo 节点信息
	NodeInfo NodeConfig `yaml:"node" json:"node"`

	// XdsResources 配置资源类型
	XdsResources ResourceConfig `yaml:"xds_resources" json:"xds_resources"`

	// Security 安全配置
	Security SecurityConfig `yaml:"security" json:"security"`

	// Backoff 重试配置
	Backoff BackoffConfig `yaml:"backoff" json:"backoff"`

	// InitialLoadTimeout 初始加载超时时间
	InitialLoadTimeout time.Duration `yaml:"initial_load_timeout" json:"initial_load_timeout"`
}

// NodeConfig 节点配置
type NodeConfig struct {
	// Id 节点唯一标识
	Id string `yaml:"id" json:"id"`

	// Cluster 集群名称
	Cluster string `yaml:"cluster" json:"cluster"`

	// Metadata 节点元数据
	Metadata map[string]string `yaml:"metadata" json:"metadata"`
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	// ListenerNames 监听器名称列表
	ListenerNames []string `yaml:"listener_names" json:"listener_names"`

	// RouteConfigNames 路由配置名称列表
	RouteConfigNames []string `yaml:"route_config_names" json:"route_config_names"`

	// ClusterNames 集群名称列表
	ClusterNames []string `yaml:"cluster_names" json:"cluster_names"`

	// Ads 是否启用ADS聚合发现服务
	Ads bool `yaml:"ads" json:"ads"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// TlsEnabled 是否启用TLS
	TlsEnabled bool `yaml:"tls_enabled" json:"tls_enabled"`

	// CaCertFile CA证书文件路径
	CaCertFile string `yaml:"ca_cert_file" json:"ca_cert_file"`

	// CertFile 客户端证书文件路径
	CertFile string `yaml:"cert_file" json:"cert_file"`

	// KeyFile 客户端私钥文件路径
	KeyFile string `yaml:"key_file" json:"key_file"`

	// ServerName TLS服务器名称
	ServerName string `yaml:"server_name" json:"server_name"`

	// InsecureSkipVerify 是否跳过TLS验证
	InsecureSkipVerify bool `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

// BackoffConfig 重试配置
type BackoffConfig struct {
	// BaseInterval 基础退避间隔
	BaseInterval time.Duration `yaml:"base_interval" json:"base_interval"`

	// MaxInterval 最大退避间隔
	MaxInterval time.Duration `yaml:"max_interval" json:"max_interval"`

	// MaxRetries 最大重试次数
	MaxRetries int `yaml:"max_retries" json:"max_retries"`
}

var (
	defaultConfig = &Config{
		ManagementServerAddresses: []string{"xds-server:15010"},
		NodeInfo: NodeConfig{
			Id:      "yggdrasil-client",
			Cluster: "yggdrasil-cluster",
		},
		XdsResources: ResourceConfig{
			ListenerNames:    []string{"*"},
			RouteConfigNames: []string{"*"},
			ClusterNames:     []string{"*"},
			Ads:              true,
		},
		Security: SecurityConfig{
			TlsEnabled:         true,
			InsecureSkipVerify: false,
		},
		Backoff: BackoffConfig{
			BaseInterval: 500 * time.Millisecond,
			MaxInterval:  30 * time.Second,
			MaxRetries:   10,
		},
		InitialLoadTimeout: 15 * time.Second,
	}
)

// SetDefaults 设置默认配置值
func (c *Config) SetDefaults() {
	if c.ManagementServerAddresses == nil {
		c.ManagementServerAddresses = []string{"localhost:15000"}
	}
	if c.NodeInfo.Id == "" {
		c.NodeInfo.Id = "default-node"
	}
	if c.NodeInfo.Cluster == "" {
		c.NodeInfo.Cluster = "default-cluster"
	}
	if c.InitialLoadTimeout == 0 {
		c.InitialLoadTimeout = 15 * time.Second
	}
	if c.Backoff.BaseInterval == 0 {
		c.Backoff.BaseInterval = 500 * time.Millisecond
	}
	if c.Backoff.MaxInterval == 0 {
		c.Backoff.MaxInterval = 30 * time.Second
	}
}
