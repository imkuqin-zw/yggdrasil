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

	resolver2 "github.com/imkuqin-zw/yggdrasil/pkg/resolver"
)

// Endpoint represents a service endpoint with xDS metadata
type Endpoint struct {
	Address  string
	Protocol string
	Metadata map[string]interface{}
}

// GetAddress returns the endpoint address
func (e *Endpoint) GetAddress() string {
	return e.Address
}

// GetProtocol returns the endpoint protocol
func (e *Endpoint) GetProtocol() string {
	return e.Protocol
}

// GetMetadata returns the endpoint metadata
func (e *Endpoint) GetMetadata() map[string]interface{} {
	return e.Metadata
}

// Ensure Endpoint implements resolver.Endpoint interface
var _ resolver2.Endpoint = (*Endpoint)(nil)

// ClusterInfo contains information about an xDS cluster
type ClusterInfo struct {
	Name      string
	Endpoints []*Endpoint
	LbPolicy  string
	Metadata  map[string]interface{}
}

// LocalityEndpoints represents endpoints grouped by locality
type LocalityEndpoints struct {
	Locality  *Locality
	Endpoints []*Endpoint
	Priority  uint32
	Weight    uint32
}

// HealthStatus represents the health status of an endpoint
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthUnhealthy
	HealthDraining
	HealthTimeout
	HealthDegraded
)

// String returns the string representation of health status
func (h HealthStatus) String() string {
	switch h {
	case HealthHealthy:
		return "HEALTHY"
	case HealthUnhealthy:
		return "UNHEALTHY"
	case HealthDraining:
		return "DRAINING"
	case HealthTimeout:
		return "TIMEOUT"
	case HealthDegraded:
		return "DEGRADED"
	default:
		return "UNKNOWN"
	}
}

// EndpointMetadata contains detailed endpoint metadata from xDS
type EndpointMetadata struct {
	// Health status
	Health HealthStatus

	// Locality information
	Locality *Locality

	// Load balancing weight
	Weight uint32

	// Priority (lower is higher priority)
	Priority uint32

	// Custom metadata
	Metadata map[string]interface{}
}

// NewEndpoint creates a new endpoint from address and metadata
func NewEndpoint(address, protocol string, metadata map[string]interface{}) *Endpoint {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return &Endpoint{
		Address:  address,
		Protocol: protocol,
		Metadata: metadata,
	}
}

// NewEndpointWithHealth creates a new endpoint with health status
func NewEndpointWithHealth(address, protocol string, health HealthStatus) *Endpoint {
	return &Endpoint{
		Address:  address,
		Protocol: protocol,
		Metadata: map[string]interface{}{
			"health": health.String(),
		},
	}
}

// GetHealth returns the health status from endpoint metadata
func (e *Endpoint) GetHealth() HealthStatus {
	if health, ok := e.Metadata["health"].(string); ok {
		switch health {
		case "HEALTHY":
			return HealthHealthy
		case "UNHEALTHY":
			return HealthUnhealthy
		case "DRAINING":
			return HealthDraining
		case "TIMEOUT":
			return HealthTimeout
		case "DEGRADED":
			return HealthDegraded
		}
	}
	return HealthUnknown
}

// IsHealthy returns true if the endpoint is healthy
func (e *Endpoint) IsHealthy() bool {
	health := e.GetHealth()
	return health == HealthHealthy || health == HealthUnknown
}

// GetLocality returns the locality from endpoint metadata
func (e *Endpoint) GetLocality() *Locality {
	if locality, ok := e.Metadata["locality"].(*Locality); ok {
		return locality
	}
	return nil
}

// GetWeight returns the load balancing weight
func (e *Endpoint) GetWeight() uint32 {
	if weight, ok := e.Metadata["weight"].(uint32); ok {
		return weight
	}
	return 1 // Default weight
}

// GetPriority returns the priority
func (e *Endpoint) GetPriority() uint32 {
	if priority, ok := e.Metadata["priority"].(uint32); ok {
		return priority
	}
	return 0 // Default priority
}

// String returns a string representation of the endpoint
func (e *Endpoint) String() string {
	return fmt.Sprintf("Endpoint{Address: %s, Protocol: %s, Health: %s}",
		e.Address, e.Protocol, e.GetHealth())
}
