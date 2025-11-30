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

//go:build integration
// +build integration

package xds

import (
	"context"
	"testing"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
)

// TestXDSIntegration tests the full xDS integration flow
// Requires: Istio Pilot running on localhost:15010
func TestXDSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup configuration
	setupTestConfig(t)

	// Test client creation
	t.Run("ClientCreation", testClientCreation)

	// Test resolver
	t.Run("Resolver", testResolver)

	// Test balancer
	t.Run("Balancer", testBalancer)

	// Cleanup
	if err := CloseClient(); err != nil {
		t.Errorf("Failed to close client: %v", err)
	}
}

func setupTestConfig(t *testing.T) {
	// Set test configuration
	testConfig := map[string]interface{}{
		"yggdrasil.xds.server.address": "localhost:15010",
		"yggdrasil.xds.node.cluster":   "test-cluster",
		"yggdrasil.xds.node.id":        "test-node-1",
		"yggdrasil.xds.resources.cds":  true,
		"yggdrasil.xds.resources.eds":  true,
		"yggdrasil.xds.timeout":        "10s",
	}

	for key, value := range testConfig {
		if err := config.Set(key, value); err != nil {
			t.Fatalf("Failed to set config %s: %v", key, err)
		}
	}
}

func testClientCreation(t *testing.T) {
	client, err := GetClient()
	if err != nil {
		t.Fatalf("Failed to create xDS client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}

	t.Log("✅ xDS client created successfully")
}

func testResolver(t *testing.T) {
	res, err := resolver.GetResolver("xds")
	if err != nil {
		t.Fatalf("Failed to get xDS resolver: %v", err)
	}

	if res.Name() != "xds" {
		t.Errorf("Expected resolver name 'xds', got '%s'", res.Name())
	}

	// Watch a test service
	serviceName := "test-service"
	if err := res.AddWatch(serviceName); err != nil {
		t.Fatalf("Failed to add watch: %v", err)
	}

	t.Logf("✅ Watching service: %s", serviceName)

	// Wait for potential endpoint updates
	time.Sleep(2 * time.Second)

	// Remove watch
	if err := res.DelWatch(serviceName); err != nil {
		t.Errorf("Failed to remove watch: %v", err)
	}

	t.Log("✅ Resolver test passed")
}

func testBalancer(t *testing.T) {
	builder := balancer.GetBuilder("xds")
	if builder == nil {
		t.Fatal("Failed to get xDS balancer builder")
	}

	serviceName := "test-service"
	bal := builder(serviceName)

	if bal == nil {
		t.Fatal("Balancer is nil")
	}

	if bal.Name() != "xds" {
		t.Errorf("Expected balancer name 'xds', got '%s'", bal.Name())
	}

	// Get picker
	picker := bal.GetPicker()
	if picker == nil {
		t.Fatal("Picker is nil")
	}

	// Try to pick an endpoint (may fail if no endpoints available)
	ctx := context.Background()
	rpcInfo := balancer.RpcInfo{
		Ctx:    ctx,
		Method: "/test.Service/Method",
	}

	result, err := picker.Next(rpcInfo)
	if err != nil {
		// This is expected if no endpoints are available
		t.Logf("No endpoints available (expected in test): %v", err)
	} else {
		endpoint := result.Endpoint()
		t.Logf("✅ Selected endpoint: %s", endpoint.GetAddress())
		result.Report(nil)
	}

	// Close balancer
	if err := bal.Close(); err != nil {
		t.Errorf("Failed to close balancer: %v", err)
	}

	t.Log("✅ Balancer test passed")
}

// TestXDSReconnection tests automatic reconnection
func TestXDSReconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setupTestConfig(t)

	client, err := GetClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Close and recreate client to test reconnection
	if err := client.Close(); err != nil {
		t.Errorf("Failed to close client: %v", err)
	}

	// Reset global client
	SetClient(nil)

	// Create new client (should reconnect)
	newClient, err := GetClient()
	if err != nil {
		t.Fatalf("Failed to reconnect: %v", err)
	}

	if newClient == nil {
		t.Fatal("New client is nil")
	}

	t.Log("✅ Reconnection test passed")

	// Cleanup
	CloseClient()
}
