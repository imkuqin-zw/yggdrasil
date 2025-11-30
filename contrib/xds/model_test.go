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
	"testing"
)

func TestEndpointHealth(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    *Endpoint
		wantHealth  HealthStatus
		wantHealthy bool
	}{
		{
			name: "healthy endpoint",
			endpoint: &Endpoint{
				Address:  "127.0.0.1:8080",
				Protocol: "grpc",
				Metadata: map[string]interface{}{
					"health": "HEALTHY",
				},
			},
			wantHealth:  HealthHealthy,
			wantHealthy: true,
		},
		{
			name: "unhealthy endpoint",
			endpoint: &Endpoint{
				Address:  "127.0.0.1:8080",
				Protocol: "grpc",
				Metadata: map[string]interface{}{
					"health": "UNHEALTHY",
				},
			},
			wantHealth:  HealthUnhealthy,
			wantHealthy: false,
		},
		{
			name: "unknown health",
			endpoint: &Endpoint{
				Address:  "127.0.0.1:8080",
				Protocol: "grpc",
				Metadata: map[string]interface{}{},
			},
			wantHealth:  HealthUnknown,
			wantHealthy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.endpoint.GetHealth(); got != tt.wantHealth {
				t.Errorf("GetHealth() = %v, want %v", got, tt.wantHealth)
			}
			if got := tt.endpoint.IsHealthy(); got != tt.wantHealthy {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.wantHealthy)
			}
		})
	}
}

func TestEndpointMetadata(t *testing.T) {
	endpoint := &Endpoint{
		Address:  "127.0.0.1:8080",
		Protocol: "grpc",
		Metadata: map[string]interface{}{
			"weight":   uint32(100),
			"priority": uint32(1),
			"locality": &Locality{
				Region:  "us-west",
				Zone:    "us-west-1a",
				SubZone: "rack-1",
			},
		},
	}

	if got := endpoint.GetWeight(); got != 100 {
		t.Errorf("GetWeight() = %v, want 100", got)
	}

	if got := endpoint.GetPriority(); got != 1 {
		t.Errorf("GetPriority() = %v, want 1", got)
	}

	locality := endpoint.GetLocality()
	if locality == nil {
		t.Fatal("GetLocality() returned nil")
	}

	if locality.Region != "us-west" {
		t.Errorf("locality.Region = %v, want us-west", locality.Region)
	}
}

func TestNewEndpoint(t *testing.T) {
	address := "127.0.0.1:8080"
	protocol := "grpc"
	metadata := map[string]interface{}{
		"test": "value",
	}

	ep := NewEndpoint(address, protocol, metadata)

	if ep.GetAddress() != address {
		t.Errorf("GetAddress() = %v, want %v", ep.GetAddress(), address)
	}

	if ep.GetProtocol() != protocol {
		t.Errorf("GetProtocol() = %v, want %v", ep.GetProtocol(), protocol)
	}

	if ep.Metadata["test"] != "value" {
		t.Error("metadata not set correctly")
	}
}

func TestHealthStatusString(t *testing.T) {
	tests := []struct {
		status HealthStatus
		want   string
	}{
		{HealthHealthy, "HEALTHY"},
		{HealthUnhealthy, "UNHEALTHY"},
		{HealthDraining, "DRAINING"},
		{HealthTimeout, "TIMEOUT"},
		{HealthDegraded, "DEGRADED"},
		{HealthUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
