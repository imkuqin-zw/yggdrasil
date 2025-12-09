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
	"errors"
	"fmt"
	"sync"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	resolver2 "github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	resolver2.RegisterBuilder(name, newResolver)
}

// resolver implements the pkg/resolver.Resolver interface using xDS EDS
type resolver struct {
	client  *Client
	mu      sync.RWMutex
	closed  bool
	watcher map[string]*watcherInstance
}

// newResolver creates a new xDS resolver
func newResolver(_ string) (resolver2.Resolver, error) {
	client, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get xDS client: %w", err)
	}

	rs := &resolver{
		client:  client,
		watcher: make(map[string]*watcherInstance),
	}

	return rs, nil
}

// AddWatch starts watching a service for endpoint updates
func (r *resolver) AddWatch(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errors.New("resolver closed")
	}

	if _, ok := r.watcher[serviceName]; ok {
		return nil // Already watching
	}

	watcher := &watcherInstance{
		serviceName: serviceName,
		resolver:    r,
	}

	// Subscribe to EDS for this service
	if err := watcher.subscribe(); err != nil {
		return err
	}

	r.watcher[serviceName] = watcher

	logger.InfoField("xDS resolver watching service",
		logger.String("service", serviceName))

	return nil
}

// DelWatch stops watching a service
func (r *resolver) DelWatch(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	w, ok := r.watcher[serviceName]
	if !ok {
		return nil
	}

	w.stop()
	delete(r.watcher, serviceName)

	logger.InfoField("xDS resolver stopped watching service",
		logger.String("service", serviceName))

	return nil
}

// Close closes the resolver and stops all watchers
func (r *resolver) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	r.mu.Unlock()

	for _, w := range r.watcher {
		w.stop()
	}

	logger.InfoField("xDS resolver closed")
	return nil
}

// Name returns the resolver name
func (r *resolver) Name() string {
	return name
}

// watcherInstance watches a specific service for endpoint updates
type watcherInstance struct {
	serviceName  string
	resolver     *resolver
	mu           sync.Mutex
	listenerName string
	routeName    string
	clusterNames []string
}

// subscribe subscribes to LDS updates for the service
// This starts the discovery chain: LDS → RDS → CDS → EDS
func (w *watcherInstance) subscribe() error {
	// Use service name as listener name
	w.listenerName = w.serviceName

	// Subscribe to LDS
	handler := func(resourceType string, resources []interface{}) error {
		return w.handleLDSUpdate(resources)
	}

	logger.DebugField("xDS resolver subscribing to LDS",
		logger.String("listener", w.listenerName))

	return w.resolver.client.Subscribe(
		resource.ListenerType,
		[]string{w.listenerName},
		handler,
	)
}

// handleLDSUpdate processes LDS updates
func (w *watcherInstance) handleLDSUpdate(resources []interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, res := range resources {
		anyRes, ok := res.(*anypb.Any)
		if !ok {
			logger.WarnField("unexpected resource type in LDS update")
			continue
		}

		// Unmarshal Listener
		lis := &listener.Listener{}
		if err := anyRes.UnmarshalTo(lis); err != nil {
			logger.ErrorField("failed to unmarshal Listener", logger.Err(err))
			continue
		}

		// Only process the listener we're watching
		if lis.Name != w.listenerName {
			continue
		}

		logger.DebugField("xDS resolver received listener",
			logger.String("listener", lis.Name))

		// Extract HTTP Connection Manager and route config
		if err := w.processListener(lis); err != nil {
			logger.ErrorField("failed to process listener",
				logger.String("listener", lis.Name),
				logger.Err(err))
			return err
		}
	}

	return nil
}

// processListener extracts route configuration from listener
func (w *watcherInstance) processListener(lis *listener.Listener) error {
	// Find HTTP Connection Manager filter
	httpConnMgr, err := w.extractHTTPConnectionManager(lis)
	if err != nil {
		return err
	}

	// Handle route configuration
	switch rds := httpConnMgr.RouteSpecifier.(type) {
	case *hcm.HttpConnectionManager_Rds:
		// RDS reference - subscribe to RDS
		routeConfigName := rds.Rds.RouteConfigName
		w.routeName = routeConfigName

		logger.DebugField("listener references RDS, subscribing",
			logger.String("listener", lis.Name),
			logger.String("route", routeConfigName))

		// Subscribe to RDS
		handler := func(resourceType string, resources []interface{}) error {
			return w.handleRDSUpdate(resources)
		}

		return w.resolver.client.Subscribe(
			resource.RouteType,
			[]string{routeConfigName},
			handler,
		)

	case *hcm.HttpConnectionManager_RouteConfig:
		// Inline route configuration
		routeConfig := rds.RouteConfig
		w.routeName = routeConfig.Name

		logger.InfoField("listener has inline route config",
			logger.String("listener", lis.Name),
			logger.String("route", routeConfig.Name))

		// Process inline route directly
		return w.handleRouteConfig(routeConfig)

	default:
		return fmt.Errorf("unsupported route specifier type")
	}
}

// extractHTTPConnectionManager extracts the HTTP Connection Manager from a listener
func (w *watcherInstance) extractHTTPConnectionManager(lis *listener.Listener) (*hcm.HttpConnectionManager, error) {
	for _, filterChain := range lis.FilterChains {
		for _, filter := range filterChain.Filters {
			if filter.Name == "envoy.filters.network.http_connection_manager" {
				httpConnMgr := &hcm.HttpConnectionManager{}

				switch c := filter.ConfigType.(type) {
				case *listener.Filter_TypedConfig:
					if err := c.TypedConfig.UnmarshalTo(httpConnMgr); err != nil {
						return nil, fmt.Errorf("failed to unmarshal HTTP connection manager: %w", err)
					}
					return httpConnMgr, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("HTTP connection manager not found in listener")
}

// handleRDSUpdate processes RDS updates
func (w *watcherInstance) handleRDSUpdate(resources []interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, res := range resources {
		anyRes, ok := res.(*anypb.Any)
		if !ok {
			logger.WarnField("unexpected resource type in RDS update")
			continue
		}

		// Unmarshal RouteConfiguration
		routeConfig := &route.RouteConfiguration{}
		if err := anyRes.UnmarshalTo(routeConfig); err != nil {
			logger.ErrorField("failed to unmarshal RouteConfiguration", logger.Err(err))
			continue
		}

		// Only process the route we're watching
		if routeConfig.Name != w.routeName {
			continue
		}

		logger.DebugField("xDS resolver received route configuration",
			logger.String("route", routeConfig.Name))

		// Process route configuration
		if err := w.handleRouteConfig(routeConfig); err != nil {
			return err
		}
	}

	return nil
}

// handleRouteConfig extracts clusters from route and subscribes to CDS
func (w *watcherInstance) handleRouteConfig(routeConfig *route.RouteConfiguration) error {
	// Extract all cluster names from the route
	clusterNames := w.extractClustersFromRoute(routeConfig)
	w.clusterNames = clusterNames

	if len(clusterNames) == 0 {
		logger.WarnField("no clusters found in route configuration",
			logger.String("route", routeConfig.Name))
		return nil
	}

	logger.DebugField("xDS resolver extracted clusters from route, subscribing to CDS",
		logger.String("route", routeConfig.Name),
		logger.Int("cluster_count", len(clusterNames)),
		logger.Any("clusters", clusterNames))

	// Subscribe to CDS for all clusters
	handler := func(resourceType string, resources []interface{}) error {
		return w.handleCDSUpdate(resources)
	}

	return w.resolver.client.Subscribe(
		resource.ClusterType,
		clusterNames,
		handler,
	)
}

// extractClustersFromRoute extracts all cluster names from a route configuration
func (w *watcherInstance) extractClustersFromRoute(r *route.RouteConfiguration) []string {
	clustersMap := make(map[string]bool)

	for _, vh := range r.VirtualHosts {
		for _, route := range vh.Routes {
			if route.GetRoute() != nil {
				// Single cluster
				if cluster := route.GetRoute().GetCluster(); cluster != "" {
					clustersMap[cluster] = true
				}

				// Weighted clusters
				if wc := route.GetRoute().GetWeightedClusters(); wc != nil {
					for _, c := range wc.Clusters {
						clustersMap[c.Name] = true
					}
				}
			}
		}
	}

	// Convert map to slice
	clusters := make([]string, 0, len(clustersMap))
	for c := range clustersMap {
		clusters = append(clusters, c)
	}

	return clusters
}

// handleCDSUpdate processes CDS updates and subscribes to EDS
func (w *watcherInstance) handleCDSUpdate(resources []interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var clustersToWatch []string

	for _, res := range resources {
		anyRes, ok := res.(*anypb.Any)
		if !ok {
			logger.WarnField("unexpected resource type in CDS update")
			continue
		}

		// Unmarshal Cluster
		cls := &cluster.Cluster{}
		if err := anyRes.UnmarshalTo(cls); err != nil {
			logger.ErrorField("failed to unmarshal Cluster", logger.Err(err))
			continue
		}

		// Check if this cluster is one we're interested in
		interested := false
		for _, cn := range w.clusterNames {
			if cls.Name == cn {
				interested = true
				break
			}
		}

		if !interested {
			continue
		}

		logger.DebugField("xDS resolver received cluster",
			logger.String("cluster", cls.Name))

		// Add to list of clusters to watch for endpoints
		clustersToWatch = append(clustersToWatch, cls.Name)
	}

	if len(clustersToWatch) == 0 {
		return nil
	}

	logger.DebugField("xDS resolver subscribing to EDS for clusters",
		logger.Int("cluster_count", len(clustersToWatch)),
		logger.Any("clusters", clustersToWatch))

	// Subscribe to EDS for these clusters
	handler := func(resourceType string, resources []interface{}) error {
		return w.handleEDSUpdate(resources)
	}

	return w.resolver.client.Subscribe(
		resource.EndpointType,
		clustersToWatch,
		handler,
	)
}

// handleEDSUpdate processes EDS updates
func (w *watcherInstance) handleEDSUpdate(resources []interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var endpoints []interface{}

	for _, res := range resources {
		anyRes, ok := res.(*anypb.Any)
		if !ok {
			logger.WarnField("unexpected resource type in EDS update")
			continue
		}

		// Unmarshal ClusterLoadAssignment
		cla := &endpoint.ClusterLoadAssignment{}
		if err := anyRes.UnmarshalTo(cla); err != nil {
			logger.ErrorField("failed to unmarshal ClusterLoadAssignment", logger.Err(err))
			continue
		}

		logger.DebugField("xDS resolver received endpoints",
			logger.String("cluster", cla.ClusterName),
			logger.Int("locality_count", len(cla.Endpoints)))

		// Extract endpoints from localities
		for _, localityEndpoints := range cla.Endpoints {
			locality := localityEndpoints.Locality
			priority := localityEndpoints.Priority
			weight := localityEndpoints.LoadBalancingWeight.GetValue()

			for _, lbEndpoint := range localityEndpoints.LbEndpoints {
				ep := lbEndpoint.GetEndpoint()
				if ep == nil {
					continue
				}

				socketAddr := ep.Address.GetSocketAddress()
				if socketAddr == nil {
					continue
				}

				address := fmt.Sprintf("%s:%d", socketAddr.GetAddress(), socketAddr.GetPortValue())

				// Determine health status
				health := HealthUnknown
				switch lbEndpoint.HealthStatus {
				case corev3.HealthStatus_HEALTHY:
					health = HealthHealthy
				case corev3.HealthStatus_UNHEALTHY:
					health = HealthUnhealthy
				case corev3.HealthStatus_DRAINING:
					health = HealthDraining
				case corev3.HealthStatus_TIMEOUT:
					health = HealthTimeout
				case corev3.HealthStatus_DEGRADED:
					health = HealthDegraded
				}

				metadata := map[string]interface{}{
					"health":   health.String(),
					"priority": priority,
					"weight":   weight,
					"cluster":  cla.ClusterName,
				}
				// Add locality information
				if locality != nil {
					metadata["locality"] = &Locality{
						Region:  locality.Region,
						Zone:    locality.Zone,
						SubZone: locality.SubZone,
					}
				}

				// Build endpoint metadata
				endpoint := map[string]interface{}{
					config2.KeySingleAddress:  address,
					config2.KeySingleProtocol: "grpc", // Default to gRPC
					config2.KeySingleMetadata: metadata,
				}

				endpoints = append(endpoints, endpoint)
			}
		}
	}

	configKey := fmt.Sprintf(config2.KeyClientEndpoints, w.serviceName)
	if err := config2.SetMulti([]string{configKey}, []interface{}{endpoints}); err != nil {
		logger.ErrorField("failed to update endpoints configuration",
			logger.String("service", w.serviceName),
			logger.Err(err))
		return err
	}

	logger.DebugField("xDS resolver updated endpoints",
		logger.String("service", w.serviceName),
		logger.Int("count", len(endpoints)))

	return nil
}

// stop stops watching the service
func (w *watcherInstance) stop() {
	// In a full implementation, we would unsubscribe from all resources here
	// For now, the client will handle cleanup when closed
}
