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
	"os"
	"sync"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// ResourceUpdateHandler is called when xDS resources are updated
type ResourceUpdateHandler func(resourceType string, resources []interface{}) error

// Client manages the connection to the xDS control plane and handles resource subscriptions
type Client struct {
	config Config
	conn   *grpc.ClientConn
	node   *core.Node

	// ADS stream
	adsClient discovery.AggregatedDiscoveryService_StreamAggregatedResourcesClient

	// Resource management
	resourceManager *ResourceManager
	ldsHandler      *LDSHandler
	rdsHandler      *RDSHandler

	// Resource handlers
	handlers  map[string][]ResourceUpdateHandler
	handlerMu sync.RWMutex

	// Subscribed resources
	subscriptions map[string]map[string]bool // resourceType -> resourceName -> subscribed
	subMu         sync.RWMutex

	// State
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed bool
	mu     sync.RWMutex
}

// NewClient creates a new xDS client
func NewClient(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize resource manager and handlers
	resourceManager := NewResourceManager()

	client := &Client{
		config:          config,
		resourceManager: resourceManager,
		handlers:        make(map[string][]ResourceUpdateHandler),
		subscriptions:   make(map[string]map[string]bool),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize handlers
	client.ldsHandler = NewLDSHandler(resourceManager, client)
	client.rdsHandler = NewRDSHandler(resourceManager)

	// Build node information
	client.node = buildNode(&config.Node)

	// Connect to xDS server
	if err := client.connect(); err != nil {
		cancel()
		return nil, err
	}

	// Start ADS stream if enabled
	if err := client.startADS(); err != nil {
		_ = client.Close()
		return nil, err
	}

	logger.InfoField("xDS client connected",
		logger.String("server", config.Server.Address),
		logger.String("node", config.Node.Id))

	return client, nil
}

// connect establishes a gRPC connection to the xDS server
func (c *Client) connect() error {
	var opts []grpc.DialOption

	// Configure TLS if enabled
	if c.config.TLS.Enabled {
		tlsConfig, err := c.buildTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add timeout
	opts = append(opts, grpc.WithBlock())

	// Determine server address
	serverAddr := c.config.Server.Address
	if c.config.Server.UseTLS && c.config.TLS.Enabled {
		// Use TLS port if specified
		serverAddr = fmt.Sprintf("%s:%d", c.config.Server.Address, c.config.Server.TLSPort)
	}

	// Connect with timeout
	ctx, cancel := context.WithTimeout(c.ctx, c.config.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	c.conn = conn
	return nil
}

// buildTLSConfig creates a TLS configuration from the config
func (c *Client) buildTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.config.TLS.InsecureSkipVerify,
	}

	// Load CA certificate
	if c.config.TLS.CACert != "" {
		caCert, err := os.ReadFile(c.config.TLS.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate if provided
	if c.config.TLS.ClientCert != "" && c.config.TLS.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(c.config.TLS.ClientCert, c.config.TLS.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Set server name for SNI
	if c.config.TLS.ServerName != "" {
		tlsConfig.ServerName = c.config.TLS.ServerName
	}

	return tlsConfig, nil
}

// buildNode creates an xDS node from configuration
func buildNode(nodeInfo *NodeInfo) *core.Node {
	node := &core.Node{
		Id:      nodeInfo.Id,
		Cluster: nodeInfo.Cluster,
	}

	// Add locality if configured
	if nodeInfo.Locality != nil {
		node.Locality = &core.Locality{
			Region:  nodeInfo.Locality.Region,
			Zone:    nodeInfo.Locality.Zone,
			SubZone: nodeInfo.Locality.SubZone,
		}
	}

	// Add metadata
	if len(nodeInfo.Metadata) > 0 {
		// Convert metadata to protobuf Struct
		// For simplicity, we'll add it as user agent name
		if buildVersion, ok := nodeInfo.Metadata["buildVersion"].(string); ok {
			node.UserAgentName = buildVersion
		}
	}

	return node
}

// startADS starts the Aggregated Discovery Service stream
func (c *Client) startADS() error {
	adsClient := discovery.NewAggregatedDiscoveryServiceClient(c.conn)

	stream, err := adsClient.StreamAggregatedResources(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to create ADS stream: %w", err)
	}

	c.adsClient = stream

	// Start goroutine to receive responses
	c.wg.Add(1)
	go c.receiveADSResponses()

	return nil
}

// receiveADSResponses receives and processes ADS responses
func (c *Client) receiveADSResponses() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		resp, err := c.adsClient.Recv()
		if err != nil {
			c.mu.RLock()
			closed := c.closed
			c.mu.RUnlock()

			if !closed {
				logger.ErrorField("failed to receive ADS response", logger.Err(err))
				// Attempt to reconnect
				c.handleDisconnect()
			}
			return
		}

		if err := c.handleADSResponse(resp); err != nil {
			logger.ErrorField("failed to handle ADS response",
				logger.String("typeUrl", resp.TypeUrl),
				logger.Err(err))
		}
	}
}

// handleADSResponse processes an ADS response
func (c *Client) handleADSResponse(resp *discovery.DiscoveryResponse) error {
	resourceType := resp.TypeUrl

	logger.DebugField("received xDS response",
		logger.String("type", resourceType),
		logger.Int("resources", len(resp.Resources)),
		logger.String("version", resp.VersionInfo))

	// Convert resources to interface slice
	var resources []interface{}
	for _, res := range resp.Resources {
		resources = append(resources, res)
	}

	// Use built-in handlers for LDS and RDS
	switch resourceType {
	case resource.ListenerType:
		if err := c.ldsHandler.HandleUpdate(resources); err != nil {
			logger.ErrorField("LDS handler failed", logger.Err(err))
		}
	case resource.RouteType:
		if err := c.rdsHandler.HandleUpdate(resources); err != nil {
			logger.ErrorField("RDS handler failed", logger.Err(err))
		}
	}

	// Call registered custom handlers
	c.handlerMu.RLock()
	handlers := c.handlers[resourceType]
	c.handlerMu.RUnlock()

	for _, handler := range handlers {
		if err := handler(resourceType, resources); err != nil {
			logger.ErrorField("handler failed",
				logger.String("type", resourceType),
				logger.Err(err))
		}
	}

	// Send ACK
	return c.sendACK(resourceType, resp.VersionInfo, resp.Nonce)
}

// sendACK sends an ACK for a received response
func (c *Client) sendACK(typeUrl, version, nonce string) error {
	c.subMu.RLock()
	resourceNames := make([]string, 0)
	if names, ok := c.subscriptions[typeUrl]; ok {
		for name := range names {
			resourceNames = append(resourceNames, name)
		}
	}
	c.subMu.RUnlock()

	req := &discovery.DiscoveryRequest{
		TypeUrl:       typeUrl,
		VersionInfo:   version,
		Node:          c.node,
		ResourceNames: resourceNames,
		ResponseNonce: nonce,
	}

	return c.adsClient.Send(req)
}

// Subscribe subscribes to a specific resource type and names
func (c *Client) Subscribe(resourceType string, resourceNames []string, handler ResourceUpdateHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClientClosed
	}
	c.mu.RUnlock()

	// Register handler
	c.handlerMu.Lock()
	c.handlers[resourceType] = append(c.handlers[resourceType], handler)
	c.handlerMu.Unlock()

	// Track subscriptions
	c.subMu.Lock()
	if c.subscriptions[resourceType] == nil {
		c.subscriptions[resourceType] = make(map[string]bool)
	}
	for _, name := range resourceNames {
		c.subscriptions[resourceType][name] = true
	}
	c.subMu.Unlock()

	// Send subscription request
	req := &discovery.DiscoveryRequest{
		TypeUrl:       resourceType,
		Node:          c.node,
		ResourceNames: resourceNames,
	}

	if c.adsClient != nil {
		if err := c.adsClient.Send(req); err != nil {
			return ErrSubscriptionFailed(resourceType, fmt.Sprintf("%v", resourceNames), err)
		}
	}

	logger.InfoField("subscribed to xDS resource",
		logger.String("type", resourceType),
		logger.Any("names", resourceNames))

	return nil
}

// handleDisconnect handles disconnection and attempts to reconnect
func (c *Client) handleDisconnect() {
	logger.WarnField("xDS connection lost, attempting to reconnect")

	ticker := time.NewTicker(c.config.RetryInterval)
	defer ticker.Stop()

	retries := 0
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.reconnect(); err != nil {
				retries++
				logger.ErrorField("reconnection failed",
					logger.Int("retries", retries),
					logger.Err(err))

				if c.config.MaxRetries > 0 && retries >= c.config.MaxRetries {
					logger.ErrorField("max retries reached, giving up")
					return
				}
			} else {
				logger.InfoField("reconnected to xDS server")
				return
			}
		}
	}
}

// reconnect attempts to reconnect to the xDS server
func (c *Client) reconnect() error {
	// Close old connection
	if c.conn != nil {
		c.conn.Close()
	}

	// Reconnect
	if err := c.connect(); err != nil {
		return err
	}

	// Restart ADS
	if err := c.startADS(); err != nil {
		return err
	}

	// Resubscribe to all resources
	c.subMu.RLock()
	for resourceType, names := range c.subscriptions {
		resourceNames := make([]string, 0, len(names))
		for name := range names {
			resourceNames = append(resourceNames, name)
		}
		c.subMu.RUnlock()

		req := &discovery.DiscoveryRequest{
			TypeUrl:       resourceType,
			Node:          c.node,
			ResourceNames: resourceNames,
		}

		if err := c.adsClient.Send(req); err != nil {
			return err
		}

		c.subMu.RLock()
	}
	c.subMu.RUnlock()

	return nil
}

// Close closes the xDS client and cleans up resources
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	logger.InfoField("closing xDS client")

	// Cancel context
	c.cancel()

	// Wait for goroutines
	c.wg.Wait()

	// Close connection
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
