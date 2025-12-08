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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"

	// Import xDS contrib module to register resolver, balancer, and registry
	_ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
)

func main() {
	// Load configuration
	if err := config.LoadFile("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.InfoField("Starting xDS example server")

	// Print xDS configuration
	printXDSConfig()

	// Create a simple HTTP server to demonstrate the service is running
	go startHealthServer()

	// Wait for shutdown signal
	waitForShutdown()

	logger.InfoField("Server shutting down")
}

func printXDSConfig() {
	xdsServerAddr := config.Get("yggdrasil.xds.server.address").String("not configured")
	xdsCluster := config.Get("yggdrasil.xds.node.cluster").String("not configured")
	xdsNodeID := config.Get("yggdrasil.xds.node.id").String("not configured")
	useTLS := config.Get("yggdrasil.xds.server.useTLS").Bool(false)

	logger.InfoField("xDS Configuration",
		logger.String("server", xdsServerAddr),
		logger.String("cluster", xdsCluster),
		logger.String("nodeId", xdsNodeID),
		logger.Bool("tls", useTLS))
}

func startHealthServer() {
	// Simple health check endpoint
	port := config.Get("server.port").Int(8080)

	logger.InfoField("Health server starting",
		logger.Int("port", port))

	// In a real implementation, you would start an actual HTTP server here
	// For this example, we'll just simulate it
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logger.InfoField("Server health check", logger.String("status", "healthy"))
	}
}

func waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logger.InfoField("Received shutdown signal", logger.String("signal", sig.String()))

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// In a real implementation, you would gracefully shutdown your services here
	select {
	case <-ctx.Done():
		logger.WarnField("Shutdown timeout exceeded")
	case <-time.After(2 * time.Second):
		logger.InfoField("Shutdown complete")
	}
}
