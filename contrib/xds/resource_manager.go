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
	"sync"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

// ResourceManager manages xDS resources (LDS, RDS, CDS, EDS) with caching and dependency tracking
type ResourceManager struct {
	mu sync.RWMutex

	// Resource caches
	listeners map[string]*listener.Listener
	routes    map[string]*route.RouteConfiguration
	clusters  map[string]*cluster.Cluster
	endpoints map[string]*endpoint.ClusterLoadAssignment

	// Dependency tracking
	// listenerName -> routeConfigName
	listenerToRoute map[string]string
	// routeConfigName -> []clusterName
	routeToClusters map[string][]string
	// clusterName -> endpointName (usually same as cluster name)
	clusterToEndpoint map[string]string
}

// NewResourceManager creates a new resource manager
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		listeners:         make(map[string]*listener.Listener),
		routes:            make(map[string]*route.RouteConfiguration),
		clusters:          make(map[string]*cluster.Cluster),
		endpoints:         make(map[string]*endpoint.ClusterLoadAssignment),
		listenerToRoute:   make(map[string]string),
		routeToClusters:   make(map[string][]string),
		clusterToEndpoint: make(map[string]string),
	}
}

// UpdateLDS updates listener resources
func (rm *ResourceManager) UpdateLDS(listeners []*listener.Listener) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, lis := range listeners {
		rm.listeners[lis.Name] = lis
		logger.DebugField("resource manager updated listener",
			logger.String("name", lis.Name))
	}

	return nil
}

// UpdateRDS updates route configuration resources
func (rm *ResourceManager) UpdateRDS(routes []*route.RouteConfiguration) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, r := range routes {
		rm.routes[r.Name] = r

		// Extract cluster dependencies
		clusters := rm.extractClustersFromRoute(r)
		rm.routeToClusters[r.Name] = clusters

		logger.DebugField("resource manager updated route",
			logger.String("name", r.Name),
			logger.Int("clusters", len(clusters)))
	}

	return nil
}

// UpdateCDS updates cluster resources
func (rm *ResourceManager) UpdateCDS(clusters []*cluster.Cluster) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, c := range clusters {
		rm.clusters[c.Name] = c

		// Track cluster to endpoint mapping
		// In most cases, endpoint name equals cluster name
		rm.clusterToEndpoint[c.Name] = c.Name

		logger.InfoField("resource manager updated cluster",
			logger.String("name", c.Name))
	}

	return nil
}

// UpdateEDS updates endpoint resources
func (rm *ResourceManager) UpdateEDS(endpoints []*endpoint.ClusterLoadAssignment) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, ep := range endpoints {
		rm.endpoints[ep.ClusterName] = ep
		logger.InfoField("resource manager updated endpoints",
			logger.String("cluster", ep.ClusterName))
	}

	return nil
}

// GetListener retrieves a listener by name
func (rm *ResourceManager) GetListener(name string) (*listener.Listener, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	lis, ok := rm.listeners[name]
	if !ok {
		return nil, fmt.Errorf("listener not found: %s", name)
	}
	return lis, nil
}

// GetRoute retrieves a route configuration by name
func (rm *ResourceManager) GetRoute(name string) (*route.RouteConfiguration, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	r, ok := rm.routes[name]
	if !ok {
		return nil, fmt.Errorf("route not found: %s", name)
	}
	return r, nil
}

// GetCluster retrieves a cluster by name
func (rm *ResourceManager) GetCluster(name string) (*cluster.Cluster, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	c, ok := rm.clusters[name]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", name)
	}
	return c, nil
}

// GetEndpoints retrieves endpoints by cluster name
func (rm *ResourceManager) GetEndpoints(clusterName string) (*endpoint.ClusterLoadAssignment, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	ep, ok := rm.endpoints[clusterName]
	if !ok {
		return nil, fmt.Errorf("endpoints not found for cluster: %s", clusterName)
	}
	return ep, nil
}

// GetClusterDependencies returns all clusters referenced by a route
func (rm *ResourceManager) GetClusterDependencies(routeName string) []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	clusters, ok := rm.routeToClusters[routeName]
	if !ok {
		return nil
	}
	return clusters
}

// GetRouteDependencies returns the route configuration referenced by a listener
func (rm *ResourceManager) GetRouteDependencies(listenerName string) string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.listenerToRoute[listenerName]
}

// SetListenerRouteMapping sets the listener to route mapping
func (rm *ResourceManager) SetListenerRouteMapping(listenerName, routeName string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.listenerToRoute[listenerName] = routeName
}

// extractClustersFromRoute extracts all cluster names from a route configuration
func (rm *ResourceManager) extractClustersFromRoute(r *route.RouteConfiguration) []string {
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
