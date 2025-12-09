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
	"sync"

	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

// Package xds provides service governance integration with xDS (Envoy's data plane API) protocol.
// It implements service discovery, load balancing, and traffic management through xDS control plane
// such as Istio Pilot, Envoy Control Plane, or other xDS-compatible servers.
//
// The package provides implementations of:
//   - Resolver: Service discovery via EDS (Endpoint Discovery Service)
//   - Balancer: Load balancing with CDS (Cluster Discovery Service)
//   - Registry: Service registration (if supported by control plane)
//
// Supported xDS APIs:
//   - LDS (Listener Discovery Service)
var (
	name = "xds"

	configKeyBase      = name
	configKeyNode      = config2.Join(configKeyBase, "node")
	configKeyServer    = config2.Join(configKeyBase, "server")
	configKeyTLS       = config2.Join(configKeyBase, "tls")
	configKeyResources = config2.Join(configKeyBase, "resources")
)

var (
	// Global xDS client instance
	globalClient *Client
	clientMu     sync.RWMutex
	clientOnce   sync.Once
)

// GetClient returns the global xDS client instance, creating it if necessary.
// The client is lazily initialized on first access.
func GetClient() (*Client, error) {
	clientMu.RLock()
	if globalClient != nil {
		clientMu.RUnlock()
		return globalClient, nil
	}
	clientMu.RUnlock()

	clientMu.Lock()
	defer clientMu.Unlock()

	// Double-check after acquiring write lock
	if globalClient != nil {
		return globalClient, nil
	}

	// Load configuration
	cfg := DefaultConfig()
	if err := config2.Scan(configKeyBase, &cfg); err != nil {
		logger.WarnField("failed to load xDS config, using defaults", logger.Err(err))
	}

	// Create client
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	globalClient = client
	return globalClient, nil
}

// SetClient sets the global xDS client instance.
// This is useful for testing or when you want to provide a custom client.
func SetClient(client *Client) {
	clientMu.Lock()
	defer clientMu.Unlock()
	globalClient = client
}

// CloseClient closes the global xDS client if it exists.
func CloseClient() error {
	clientMu.Lock()
	defer clientMu.Unlock()

	if globalClient != nil {
		err := globalClient.Close()
		globalClient = nil
		return err
	}
	return nil
}

// namespace returns the xDS namespace, defaulting to "default" if not configured
func namespace(ns string) string {
	if ns == "" {
		return "default"
	}
	return ns
}

// buildNodeInfo creates node information from configuration
func buildNodeInfo() *NodeInfo {
	node := &NodeInfo{}
	if err := config2.Scan(configKeyNode, node); err != nil {
		logger.WarnField("failed to load xDS node config", logger.Err(err))
	}

	// Set defaults if not configured
	if node.Cluster == "" {
		node.Cluster = config2.Get("yggdrasil.cluster").String("default-cluster")
	}
	if node.Id == "" {
		node.Id = config2.Get("yggdrasil.instance.id").String("")
	}

	return node
}
