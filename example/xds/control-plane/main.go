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
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"gopkg.in/yaml.v3"

	"github.com/imkuqin-zw/yggdrasil/example/xds/control-plane/pkg/server"
	"github.com/imkuqin-zw/yggdrasil/example/xds/control-plane/pkg/snapshot"
	"github.com/imkuqin-zw/yggdrasil/example/xds/control-plane/pkg/watcher"
)

// Config represents the server configuration
type Config struct {
	Server struct {
		Port   uint   `yaml:"port"`
		NodeID string `yaml:"nodeID"`
	} `yaml:"server"`
	XDS struct {
		ConfigFile    string `yaml:"configFile"`
		WatchInterval string `yaml:"watchInterval"`
	} `yaml:"xds"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

var (
	snapshotVersion atomic.Uint64
	nodeID          = "yggdrasil.example.xds.client.1"
)

func main() {
	log.Println("Starting xDS Control Plane Server...")

	// Load server configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Server configuration loaded: port=%d, configFile=%s", config.Server.Port, config.XDS.ConfigFile)

	// Create snapshot cache
	snapshotCache := cache.NewSnapshotCache(false, cache.IDHash{}, nil)

	// Load initial xDS configuration
	if err := loadAndUpdateSnapshot(config.XDS.ConfigFile, snapshotCache); err != nil {
		log.Fatalf("Failed to load initial xDS configuration: %v", err)
	}

	// Setup file watcher for dynamic updates
	watchInterval := parseDuration(config.XDS.WatchInterval, 1*time.Second)
	fw, err := watcher.NewFileWatcher(config.XDS.ConfigFile, func(filePath string) {
		log.Printf("Configuration file changed, reloading: %s", filePath)
		if err := loadAndUpdateSnapshot(filePath, snapshotCache); err != nil {
			log.Printf("Failed to reload configuration: %v", err)
		}
	}, watchInterval)
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}
	defer fw.Close()

	fw.Start()
	log.Printf("Watching configuration file: %s", config.XDS.ConfigFile)

	// Create and start xDS server
	xdsServer := server.NewServer(config.Server.Port, snapshotCache)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := xdsServer.Run(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("Received shutdown signal: %s", sig)
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// Graceful shutdown
	log.Println("Shutting down xDS server...")
	xdsServer.Stop()

	// Give some time for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
	log.Println("xDS Control Plane Server stopped")
}

// loadConfig loads the server configuration from a YAML file
func loadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// loadAndUpdateSnapshot loads xDS configuration and updates the snapshot cache
func loadAndUpdateSnapshot(filePath string, snapshotCache cache.SnapshotCache) error {
	// Read configuration file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read xDS config file: %w", err)
	}

	// Parse configuration
	var xdsConfig snapshot.XDSConfig
	if err := yaml.Unmarshal(data, &xdsConfig); err != nil {
		return fmt.Errorf("failed to parse xDS config: %w", err)
	}

	// Increment version
	version := snapshotVersion.Add(1)
	versionStr := strconv.FormatUint(version, 10)

	log.Printf("Building snapshot version %s with %d clusters, %d endpoints, %d listeners, %d routes",
		versionStr, len(xdsConfig.Clusters), len(xdsConfig.Endpoints), len(xdsConfig.Listeners), len(xdsConfig.Routes))

	// Build snapshot
	builder := snapshot.NewBuilder(versionStr)
	snap, err := builder.BuildSnapshot(&xdsConfig)
	if err != nil {
		return fmt.Errorf("failed to build snapshot: %w", err)
	}

	// Validate snapshot
	if err := snap.Consistent(); err != nil {
		return fmt.Errorf("snapshot inconsistency: %w", err)
	}

	// Update cache
	if err := snapshotCache.SetSnapshot(context.Background(), nodeID, snap); err != nil {
		return fmt.Errorf("failed to set snapshot: %w", err)
	}

	log.Printf("Snapshot version %s successfully updated in cache", versionStr)
	return nil
}

// parseDuration parses a duration string with a default fallback
func parseDuration(s string, defaultDuration time.Duration) time.Duration {
	if s == "" {
		return defaultDuration
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("Failed to parse duration '%s', using default %v: %v", s, defaultDuration, err)
		return defaultDuration
	}
	return d
}
