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

	// Resilience components per cluster
	circuitBreakers  map[string]*CircuitBreaker
	outlierDetectors map[string]*OutlierDetector
	rateLimiters     map[string]*RateLimiter
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
		circuitBreakers:  make(map[string]*CircuitBreaker),
		outlierDetectors: make(map[string]*OutlierDetector),
		rateLimiters:     make(map[string]*RateLimiter),
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

		clusterInfo := &ClusterInfo{
			Name:     c.Name,
			LbPolicy: lbPolicy,
			Metadata: make(map[string]interface{}),
		}

		// Parse circuit breaker thresholds
		if c.CircuitBreakers != nil && len(c.CircuitBreakers.Thresholds) > 0 {
			threshold := c.CircuitBreakers.Thresholds[0]
			clusterInfo.CircuitBreaker = &CircuitBreakerConfig{
				MaxConnections:     threshold.MaxConnections.GetValue(),
				MaxPendingRequests: threshold.MaxPendingRequests.GetValue(),
				MaxRequests:        threshold.MaxRequests.GetValue(),
				MaxRetries:         threshold.MaxRetries.GetValue(),
			}

			// Initialize circuit breaker
			b.circuitBreakers[c.Name] = NewCircuitBreaker(clusterInfo.CircuitBreaker)

			logger.InfoField("circuit breaker configured for cluster",
				logger.String("cluster", c.Name),
				logger.Uint32("maxConnections", clusterInfo.CircuitBreaker.MaxConnections),
				logger.Uint32("maxRequests", clusterInfo.CircuitBreaker.MaxRequests))
		}

		// Parse outlier detection configuration
		if c.OutlierDetection != nil {
			od := c.OutlierDetection
			clusterInfo.OutlierDetection = &OutlierDetectionConfig{
				Consecutive5xx:                 od.Consecutive_5Xx.GetValue(),
				ConsecutiveGatewayFailure:      od.ConsecutiveGatewayFailure.GetValue(),
				ConsecutiveLocalOriginFailure:  od.ConsecutiveLocalOriginFailure.GetValue(),
				Interval:                       od.Interval.AsDuration(),
				BaseEjectionTime:               od.BaseEjectionTime.AsDuration(),
				MaxEjectionTime:                od.MaxEjectionTime.AsDuration(),
				MaxEjectionPercent:             od.MaxEjectionPercent.GetValue(),
				EnforcingConsecutive5xx:        od.EnforcingConsecutive_5Xx.GetValue(),
				EnforcingSuccessRate:           od.EnforcingSuccessRate.GetValue(),
				SuccessRateMinimumHosts:        od.SuccessRateMinimumHosts.GetValue(),
				SuccessRateRequestVolume:       od.SuccessRateRequestVolume.GetValue(),
				SuccessRateStdevFactor:         od.SuccessRateStdevFactor.GetValue(),
				FailurePercentageThreshold:     od.FailurePercentageThreshold.GetValue(),
				EnforcingFailurePercentage:     od.EnforcingFailurePercentage.GetValue(),
				FailurePercentageMinimumHosts:  od.FailurePercentageMinimumHosts.GetValue(),
				FailurePercentageRequestVolume: od.FailurePercentageRequestVolume.GetValue(),
				SplitExternalLocalOriginErrors: od.SplitExternalLocalOriginErrors,
			}

			// Initialize outlier detector
			// Stop existing detector if present
			if oldDetector, ok := b.outlierDetectors[c.Name]; ok {
				oldDetector.Stop()
			}

			detector := NewOutlierDetector(clusterInfo.OutlierDetection)
			b.outlierDetectors[c.Name] = detector
			detector.Start()

			logger.InfoField("outlier detection configured for cluster",
				logger.String("cluster", c.Name),
				logger.Uint32("consecutive5xx", clusterInfo.OutlierDetection.Consecutive5xx),
				logger.Duration("interval", clusterInfo.OutlierDetection.Interval))
		}

		// Parse rate limit from cluster metadata
		if metadata := c.Metadata; metadata != nil {
			if filterMetadata, ok := metadata.FilterMetadata["yggdrasil.rate_limit"]; ok {
				if fields := filterMetadata.Fields; fields != nil {
					rlConfig := &RateLimitConfig{}
					if v, ok := fields["max_tokens"]; ok {
						rlConfig.MaxTokens = uint32(v.GetNumberValue())
					}
					if v, ok := fields["tokens_per_fill"]; ok {
						rlConfig.TokensPerFill = uint32(v.GetNumberValue())
					}
					if v, ok := fields["fill_interval"]; ok {
						rlConfig.FillInterval = time.Duration(v.GetNumberValue()) * time.Second
					}

					if rlConfig.MaxTokens > 0 {
						clusterInfo.RateLimiter = rlConfig

						// Stop existing rate limiter if present
						if oldLimiter, ok := b.rateLimiters[c.Name]; ok {
							oldLimiter.Stop()
						}

						limiter := NewRateLimiter(rlConfig)
						b.rateLimiters[c.Name] = limiter

						logger.InfoField("rate limit configured for cluster",
							logger.String("cluster", c.Name),
							logger.Uint32("maxTokens", rlConfig.MaxTokens))
					}
				}
			}
		}

		b.clusters[c.Name] = clusterInfo

		logger.DebugField("xDS balancer updated cluster config",
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
	b.mu.Lock()
	defer b.mu.Unlock()

	// Stop all outlier detectors
	for _, od := range b.outlierDetectors {
		od.Stop()
	}

	// Stop all rate limiters
	for _, rl := range b.rateLimiters {
		rl.Stop()
	}

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

	// Check rate limiter first
	p.balancer.mu.RLock()
	rateLimiter := p.balancer.rateLimiters[clusterName]
	p.balancer.mu.RUnlock()

	if rateLimiter != nil && !rateLimiter.Allow() {
		return nil, ErrRateLimitExceeded
	}

	// Try to acquire circuit breaker slot
	p.balancer.mu.RLock()
	circuitBreaker := p.balancer.circuitBreakers[clusterName]
	p.balancer.mu.RUnlock()

	if circuitBreaker != nil {
		if !circuitBreaker.TryAcquire(ResourceRequest) {
			return nil, ErrCircuitBreakerOpen
		}
	}

	// Get endpoints for the selected cluster
	p.balancer.mu.RLock()
	endpoints := p.balancer.clusterEndpoints[clusterName]
	clusterInfo := p.balancer.clusters[clusterName]
	outlierDetector := p.balancer.outlierDetectors[clusterName]
	p.balancer.mu.RUnlock()

	if len(endpoints) == 0 {
		// Release circuit breaker slot if acquired
		if circuitBreaker != nil {
			circuitBreaker.Release(ResourceRequest)
		}
		return nil, balancer2.ErrNoAvailableInstance
	}

	// Filter healthy endpoints (not ejected by outlier detection)
	healthyEndpoints := make([]*Endpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		if ep.IsHealthy() {
			// Check if endpoint is ejected by outlier detection
			if outlierDetector != nil && outlierDetector.IsEjected(ep.Address) {
				continue
			}
			healthyEndpoints = append(healthyEndpoints, ep)
		}
	}

	if len(healthyEndpoints) == 0 {
		// Release circuit breaker slot if acquired
		if circuitBreaker != nil {
			circuitBreaker.Release(ResourceRequest)
		}
		return nil, balancer2.ErrNoAvailableInstance
	}

	// Select endpoint based on load balancing policy
	lbPolicy := "round_robin"
	if clusterInfo != nil {
		lbPolicy = clusterInfo.LbPolicy
	}

	selectedEndpoint := p.selectEndpoint(healthyEndpoints, lbPolicy)

	result := &pickResult{
		endpoint:        selectedEndpoint,
		circuitBreaker:  circuitBreaker,
		outlierDetector: outlierDetector,
		report: func(err error) {
			p.reportResult(selectedEndpoint, circuitBreaker, outlierDetector, err)
		},
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
func (p *picker) reportResult(endpoint *Endpoint, cb *CircuitBreaker, od *OutlierDetector, err error) {
	// Release circuit breaker slot
	if cb != nil {
		cb.Release(ResourceRequest)
	}

	// Report to outlier detector
	if od != nil {
		// Extract status code from error if available
		statusCode := 200
		if err != nil {
			// Default to 500 for errors
			statusCode = 500
			// TODO: Extract actual status code from error if available
		}
		od.ReportResult(endpoint.Address, err, statusCode)
	}

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
	endpoint        *Endpoint
	circuitBreaker  *CircuitBreaker
	outlierDetector *OutlierDetector
	report          func(err error)
}

// Endpoint returns the selected endpoint
func (p *pickResult) Endpoint() resolver2.Endpoint {
	return p.endpoint
}

// Report reports the result of using this endpoint
func (p *pickResult) Report(err error) {
	p.report(err)
}
