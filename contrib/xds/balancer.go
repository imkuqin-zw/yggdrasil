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
	"math/rand"
	"strings"
	"sync"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	balancer2 "github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	resolver2 "github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	balancer2.RegisterBuilder(name, newBalancer)
}

// balancer implements the pkg/balancer.Balancer interface with xDS routing
type balancer struct {
	serviceName string
	client      *Client
	mu          sync.RWMutex

	// Resource manager and route matcher
	resourceManager *ResourceManager
	routeMatcher    *RouteMatcher

	// Cluster configurations
	clusters map[string]*ClusterInfo // clusterName -> cluster info

	// Endpoints per cluster
	clusterEndpoints map[string][]*Endpoint // clusterName -> endpoints
}

// newBalancer creates a new xDS balancer
func newBalancer(serviceName string) balancer2.Balancer {
	client, err := GetClient()
	if err != nil {
		logger.ErrorField("failed to get xDS client", logger.Err(err))
		return nil
	}

	b := &balancer{
		serviceName:      serviceName,
		client:           client,
		resourceManager:  client.resourceManager,
		routeMatcher:     NewRouteMatcher(client.resourceManager),
		clusters:         make(map[string]*ClusterInfo),
		clusterEndpoints: make(map[string][]*Endpoint),
	}

	// Subscribe to LDS for this service
	ldsHandler := func(resourceType string, resources []interface{}) error {
		// LDS updates are handled by client's built-in handler
		return nil
	}
	if err := client.Subscribe(resource.ListenerType, []string{serviceName}, ldsHandler); err != nil {
		logger.ErrorField("failed to subscribe to LDS",
			logger.String("service", serviceName),
			logger.Err(err))
	}

	// Subscribe to CDS for cluster configuration
	cdsHandler := func(resourceType string, resources []interface{}) error {
		return b.handleCDSUpdate(resources)
	}
	if err := client.Subscribe(resource.ClusterType, []string{serviceName}, cdsHandler); err != nil {
		logger.ErrorField("failed to subscribe to CDS",
			logger.String("service", serviceName),
			logger.Err(err))
	}

	return b
}

// handleCDSUpdate processes CDS updates
func (b *balancer) handleCDSUpdate(resources []interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, res := range resources {
		anyRes, ok := res.(*anypb.Any)
		if !ok {
			continue
		}

		// Unmarshal Cluster
		c := &cluster.Cluster{}
		if err := anyRes.UnmarshalTo(c); err != nil {
			logger.ErrorField("failed to unmarshal Cluster", logger.Err(err))
			continue
		}

		// Extract load balancing policy
		lbPolicy := "round_robin"
		if c.LbPolicy == cluster.Cluster_LEAST_REQUEST {
			lbPolicy = "least_request"
		} else if c.LbPolicy == cluster.Cluster_RANDOM {
			lbPolicy = "random"
		}

		b.clusters[c.Name] = &ClusterInfo{
			Name:     c.Name,
			LbPolicy: lbPolicy,
			Metadata: make(map[string]interface{}),
		}

		logger.InfoField("xDS balancer updated cluster config",
			logger.String("cluster", c.Name),
			logger.String("lbPolicy", lbPolicy))
	}

	return nil
}

// GetPicker returns a picker for endpoint selection
func (b *balancer) GetPicker() balancer2.Picker {
	return &picker{
		balancer: b,
	}
}

// Update updates the balancer configuration with endpoints
func (b *balancer) Update(cfg config2.Values) {
	// Parse endpoints from config
	type endpointConfig struct {
		Address  string                 `yaml:"address"`
		Protocol string                 `yaml:"protocol"`
		Metadata map[string]interface{} `yaml:"metadata"`
	}

	var endpoints []endpointConfig
	if err := cfg.Get(config2.KeySingleEndpoints).Scan(&endpoints); err != nil {
		logger.ErrorField("failed to scan endpoints from config",
			logger.String("service", b.serviceName),
			logger.Err(err))
		return
	}

	// Convert to internal Endpoint type
	b.mu.Lock()
	defer b.mu.Unlock()

	// Group endpoints by cluster (extract from metadata if available)
	for _, epCfg := range endpoints {
		ep := &Endpoint{
			Address:  epCfg.Address,
			Protocol: epCfg.Protocol,
			Metadata: epCfg.Metadata,
		}
		if ep.Metadata == nil {
			ep.Metadata = make(map[string]interface{})
		}

		// Determine which cluster this endpoint belongs to
		// Default to service name if not specified
		clusterName := b.serviceName
		if cluster, ok := ep.Metadata["cluster"].(string); ok {
			clusterName = cluster
		}

		b.clusterEndpoints[clusterName] = append(b.clusterEndpoints[clusterName], ep)
	}

	logger.DebugField("xDS balancer configuration updated",
		logger.String("service", b.serviceName),
		logger.Int("clusters", len(b.clusterEndpoints)))
}

// Close closes the balancer
func (b *balancer) Close() error {
	logger.InfoField("xDS balancer closed", logger.String("service", b.serviceName))
	return nil
}

// Name returns the balancer name
func (b *balancer) Name() string {
	return name
}

// picker implements the balancer.Picker interface with routing
type picker struct {
	balancer *balancer
	mu       sync.Mutex
	index    int
	rng      *rand.Rand
}

// Next selects the next endpoint based on routing rules
func (p *picker) Next(ri balancer2.RpcInfo) (balancer2.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Extract request information from RpcInfo
	req := p.extractRequest(ri)

	// Match route to determine target cluster
	clusterName, err := p.selectCluster(req)
	if err != nil {
		// Fallback to service name as cluster
		clusterName = p.balancer.serviceName
		logger.DebugField("using default cluster",
			logger.String("cluster", clusterName),
			logger.Err(err))
	}

	// Get endpoints for the selected cluster
	p.balancer.mu.RLock()
	endpoints := p.balancer.clusterEndpoints[clusterName]
	clusterInfo := p.balancer.clusters[clusterName]
	p.balancer.mu.RUnlock()

	if len(endpoints) == 0 {
		return nil, balancer2.ErrNoAvailableInstance
	}

	// Filter healthy endpoints
	healthyEndpoints := make([]*Endpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		if ep.IsHealthy() {
			healthyEndpoints = append(healthyEndpoints, ep)
		}
	}

	if len(healthyEndpoints) == 0 {
		return nil, balancer2.ErrNoAvailableInstance
	}

	// Select endpoint based on load balancing policy
	lbPolicy := "round_robin"
	if clusterInfo != nil {
		lbPolicy = clusterInfo.LbPolicy
	}

	selectedEndpoint := p.selectEndpoint(healthyEndpoints, lbPolicy)

	result := &pickResult{
		endpoint: selectedEndpoint,
		report:   func(err error) { p.reportResult(selectedEndpoint, err) },
	}

	return result, nil
}

// extractRequest extracts request information from RpcInfo
func (p *picker) extractRequest(ri balancer2.RpcInfo) *Request {
	req := &Request{
		Method:  ri.Method,
		Headers: make(map[string]string),
	}

	// Extract path from method (format: /service/method)
	parts := strings.Split(ri.Method, "/")
	if len(parts) >= 2 {
		req.Path = "/" + strings.Join(parts[1:], "/")
	} else {
		req.Path = ri.Method
	}

	// Extract headers from context if available
	// This depends on the framework's RpcInfo implementation
	// For now, we'll use basic extraction
	if ctx := ri.Ctx; ctx != nil {
		// Extract headers from context
		md, ok := metadata.FromOutContext(ctx)
		if ok {
			req.Headers = make(map[string]string, len(md))
			for k, v := range md {
				req.Headers[k] = strings.Join(v, ";")
			}
		}
	}

	// Extract host
	req.Host = p.balancer.serviceName

	return req
}

// selectCluster selects a cluster based on routing rules
func (p *picker) selectCluster(req *Request) (string, error) {
	// Get route configuration for this service
	routeConfigName := p.balancer.resourceManager.GetRouteDependencies(p.balancer.serviceName)
	if routeConfigName == "" {
		// No route config, use service name as cluster
		return p.balancer.serviceName, nil
	}

	// Match route
	match, err := p.balancer.routeMatcher.Match(routeConfigName, req)
	if err != nil {
		return "", err
	}

	// Handle weighted clusters
	if match.WeightedClusters != nil {
		return SelectWeightedCluster(match.WeightedClusters), nil
	}

	// Handle cluster header
	if match.ClusterHeader != "" {
		if cluster, ok := req.Headers[match.ClusterHeader]; ok {
			return cluster, nil
		}
	}

	// Return single cluster
	return match.Cluster, nil
}

// selectEndpoint selects an endpoint based on load balancing policy
func (p *picker) selectEndpoint(endpoints []*Endpoint, lbPolicy string) *Endpoint {
	switch lbPolicy {
	case "round_robin":
		selected := endpoints[p.index%len(endpoints)]
		p.index++
		return selected

	case "random":
		if p.rng == nil {
			p.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		}
		return endpoints[p.rng.Intn(len(endpoints))]

	case "least_request":
		// For simplicity, use round robin for now
		// A full implementation would track active requests per endpoint
		selected := endpoints[p.index%len(endpoints)]
		p.index++
		return selected

	default:
		selected := endpoints[p.index%len(endpoints)]
		p.index++
		return selected
	}
}

// reportResult reports the result of an RPC call
func (p *picker) reportResult(endpoint *Endpoint, err error) {
	if err != nil {
		logger.WarnField("RPC call failed",
			logger.String("endpoint", endpoint.Address),
			logger.Err(err))
	} else {
		logger.DebugField("RPC call succeeded",
			logger.String("endpoint", endpoint.Address))
	}
}

// pickResult implements the balancer.PickResult interface
type pickResult struct {
	endpoint *Endpoint
	report   func(err error)
}

// Endpoint returns the selected endpoint
func (p *pickResult) Endpoint() resolver2.Endpoint {
	return p.endpoint
}

// Report reports the result of using this endpoint
func (p *pickResult) Report(err error) {
	p.report(err)
}
