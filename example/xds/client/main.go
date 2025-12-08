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

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"

	// Import xDS contrib module
	_ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)

func main() {
	// Load configuration
	if err := config.LoadFile("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.InfoField("Starting xDS example client")

	// Demonstrate service discovery
	demonstrateServiceDiscovery()

	// Demonstrate load balancing
	demonstrateLoadBalancing()

	logger.InfoField("Client demo complete")
}

func demonstrateServiceDiscovery() {
	logger.InfoField("=== Service Discovery Demo ===")

	// Get the xDS resolver
	res, err := resolver.GetResolver("xds")
	if err != nil {
		logger.ErrorField("Failed to get xDS resolver", logger.Err(err))
		return
	}

	// Watch a service
	serviceName := "example-service"
	if err := res.AddWatch(serviceName); err != nil {
		logger.ErrorField("Failed to watch service",
			logger.String("service", serviceName),
			logger.Err(err))
		return
	}

	logger.InfoField("Watching service via xDS",
		logger.String("service", serviceName))

	// Wait for endpoints to be discovered
	time.Sleep(5 * time.Second)

	// Check if endpoints were discovered
	endpointsKey := fmt.Sprintf("yggdrasil.client.%s.endpoints", serviceName)
	endpoints := config.Get(endpointsKey).Value()

	if endpoints != nil {
		logger.InfoField("Discovered endpoints",
			logger.String("service", serviceName),
			logger.Any("endpoints", endpoints))
	} else {
		logger.WarnField("No endpoints discovered yet",
			logger.String("service", serviceName))
	}
}

func demonstrateLoadBalancing() {
	logger.InfoField("=== Load Balancing Demo ===")

	// Get the xDS balancer
	bal := balancer.GetBuilder("xds")
	if bal == nil {
		logger.ErrorField("Failed to get xDS balancer builder")
		return
	}

	// Create balancer for a service
	serviceName := "example-service"
	serviceBalancer := bal(serviceName)

	logger.InfoField("Created xDS balancer",
		logger.String("service", serviceName),
		logger.String("balancer", serviceBalancer.Name()))

	// Get picker
	picker := serviceBalancer.GetPicker()

	// Simulate multiple requests
	for i := 0; i < 5; i++ {
		ctx := context.Background()
		rpcInfo := balancer.RpcInfo{
			Ctx:    ctx,
			Method: "/example.Service/Method",
		}

		result, err := picker.Next(rpcInfo)
		if err != nil {
			logger.ErrorField("Failed to pick endpoint",
				logger.Int("attempt", i+1),
				logger.Err(err))
			continue
		}

		endpoint := result.Endpoint()
		logger.InfoField("Selected endpoint",
			logger.Int("attempt", i+1),
			logger.String("address", endpoint.GetAddress()),
			logger.String("protocol", endpoint.GetProtocol()))

		// Simulate successful call
		result.Report(nil)

		time.Sleep(500 * time.Millisecond)
	}
}
