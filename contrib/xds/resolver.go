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

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
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
	serviceName string
	resolver    *resolver
	mu          sync.Mutex
}

// subscribe subscribes to EDS updates for the service
func (w *watcherInstance) subscribe() error {
	// Cluster name is typically the service name in Istio
	clusterName := w.serviceName

	// Subscribe to EDS
	handler := func(resourceType string, resources []interface{}) error {
		return w.handleEDSUpdate(resources)
	}

	return w.resolver.client.Subscribe(
		resource.EndpointType,
		[]string{clusterName},
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

				// Build endpoint metadata
				metadata := map[string]interface{}{
					config2.KeySingleAddress:  address,
					config2.KeySingleProtocol: "grpc", // Default to gRPC
					"health":                  health.String(),
					"priority":                priority,
					"weight":                  weight,
				}

				// Add locality information
				if locality != nil {
					metadata["locality"] = &Locality{
						Region:  locality.Region,
						Zone:    locality.Zone,
						SubZone: locality.SubZone,
					}
				}

				endpoints = append(endpoints, metadata)
			}
		}
	}

	// Update configuration with new endpoints
	// endpoints is already in the correct format: []interface{} containing map[string]interface{}
	configKey := fmt.Sprintf(config2.KeyClientEndpoints, w.serviceName)
	if err := config2.Set(configKey, endpoints); err != nil {
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
	// In a full implementation, we would unsubscribe from EDS here
	// For now, the client will handle cleanup when closed
}
