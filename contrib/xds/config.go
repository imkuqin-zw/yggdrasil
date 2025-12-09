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
	"time"
)

// Config contains all xDS client configuration
type Config struct {
	// Server configuration
	Server ServerConfig `yaml:"server" json:"server"`

	// Node identification
	Node NodeInfo `yaml:"node" json:"node"`

	// TLS configuration
	TLS TLSConfig `yaml:"tls" json:"tls"`

	// Resources to subscribe to
	Resources ResourcesConfig `yaml:"resources" json:"resources"`

	// Connection settings
	Timeout       time.Duration `yaml:"timeout" json:"timeout" default:"10s"`
	RetryInterval time.Duration `yaml:"retryInterval" json:"retryInterval" default:"5s"`
	MaxRetries    int           `yaml:"maxRetries" json:"maxRetries" default:"3"`
}

// ServerConfig contains xDS server connection settings
type ServerConfig struct {
	// Address of the xDS server (e.g., "localhost:15010")
	Address string `yaml:"address" json:"address" default:"localhost:15010"`

	// UseTLS indicates whether to use TLS connection
	UseTLS bool `yaml:"useTLS" json:"useTLS" default:"false"`

	// TLS port (used when UseTLS is true)
	TLSPort int `yaml:"tlsPort" json:"tlsPort" default:"15011"`
}

// NodeInfo contains node identification information sent to xDS server
type NodeInfo struct {
	// Cluster name
	Cluster string `yaml:"cluster" json:"cluster" default:"default-cluster"`

	// Node ID (unique identifier for this instance)
	Id string `yaml:"id" json:"id"`

	// Locality information
	Locality *Locality `yaml:"locality" json:"locality"`

	// Metadata attached to the node
	Metadata map[string]interface{} `yaml:"metadata" json:"metadata"`

	// Build version
	BuildVersion string `yaml:"buildVersion" json:"buildVersion" default:"1.19.0"`
}

// Locality represents the location of the node
type Locality struct {
	Region  string `yaml:"region" json:"region"`
	Zone    string `yaml:"zone" json:"zone"`
	SubZone string `yaml:"subZone" json:"subZone"`
}

// TLSConfig contains TLS/SSL configuration
type TLSConfig struct {
	// Enable TLS
	Enabled bool `yaml:"enabled" json:"enabled" default:"false"`

	// Path to CA certificate file
	CACert string `yaml:"caCert" json:"caCert"`

	// Path to client certificate file
	ClientCert string `yaml:"clientCert" json:"clientCert"`

	// Path to client private key file
	ClientKey string `yaml:"clientKey" json:"clientKey"`

	// Server name for SNI
	ServerName string `yaml:"serverName" json:"serverName"`

	// Skip certificate verification (insecure, for testing only)
	InsecureSkipVerify bool `yaml:"insecureSkipVerify" json:"insecureSkipVerify" default:"false"`
}

// ResourcesConfig specifies which xDS resources to subscribe to
type ResourcesConfig struct {
	// Subscribe to LDS (Listener Discovery Service)
	LDS bool `yaml:"lds" json:"lds" default:"true"`

	// Subscribe to RDS (Route Discovery Service)
	RDS bool `yaml:"rds" json:"rds" default:"true"`

	// Subscribe to CDS (Cluster Discovery Service)
	CDS bool `yaml:"cds" json:"cds" default:"true"`

	// Subscribe to EDS (Endpoint Discovery Service)
	EDS bool `yaml:"eds" json:"eds" default:"true"`

	// Specific resource names to watch (empty means watch all)
	ClusterNames  []string `yaml:"clusterNames" json:"clusterNames"`
	ListenerNames []string `yaml:"listenerNames" json:"listenerNames"`
	RouteNames    []string `yaml:"routeNames" json:"routeNames"`
}

// DefaultConfig returns a default xDS configuration suitable for Istio Pilot
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Address: "localhost:15010",
			UseTLS:  false,
			TLSPort: 15011,
		},
		Node: NodeInfo{
			Cluster:      "default-cluster",
			BuildVersion: "1.19.0",
			Metadata:     make(map[string]interface{}),
		},
		TLS: TLSConfig{
			Enabled:            false,
			InsecureSkipVerify: false,
		},
		Resources: ResourcesConfig{
			LDS: true,
			RDS: true,
			CDS: true,
			EDS: true,
		},
		Timeout:       10 * time.Second,
		RetryInterval: 5 * time.Second,
		MaxRetries:    3,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return ErrInvalidConfig("server address is required")
	}
	if c.Node.Cluster == "" {
		return ErrInvalidConfig("node cluster is required")
	}
	if c.TLS.Enabled {
		if c.TLS.CACert == "" {
			return ErrInvalidConfig("CA certificate is required when TLS is enabled")
		}
	}
	return nil
}

// GetServerAddress returns the appropriate server address based on TLS settings
func (c *Config) GetServerAddress() string {
	if c.Server.UseTLS && c.TLS.Enabled {
		// Parse host from address and use TLS port
		return c.Server.Address // Will be handled in client connection logic
	}
	return c.Server.Address
}
