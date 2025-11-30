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
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	registry2 "github.com/imkuqin-zw/yggdrasil/pkg/registry"
)

func init() {
	registry2.RegisterBuilder(name, buildRegistry)
}

// RegistryConfig contains xDS registry configuration
type RegistryConfig struct {
	// Enable service registration (not all xDS control planes support this)
	EnableRegistration bool `yaml:"enableRegistration" json:"enableRegistration" default:"false"`

	// Service token for authentication
	ServiceToken string `yaml:"serviceToken" json:"serviceToken"`
}

// registry implements the pkg/registry.Registry interface
// Note: xDS is primarily for service discovery, not registration
// Registration support depends on the control plane implementation
type registry struct {
	config RegistryConfig
	client *Client
}

// buildRegistry creates a new xDS registry
func buildRegistry() registry2.Registry {
	cfg := RegistryConfig{}
	// Load configuration if needed

	client, err := GetClient()
	if err != nil {
		logger.ErrorField("failed to get xDS client for registry", logger.Err(err))
		return nil
	}

	return &registry{
		config: cfg,
		client: client,
	}
}

// Register registers a service instance
// Note: Most xDS control planes (like Istio) handle registration automatically
// through service discovery mechanisms (e.g., Kubernetes service discovery)
func (r *registry) Register(ctx context.Context, info registry2.Instance) error {
	if !r.config.EnableRegistration {
		logger.InfoField("xDS service registration is disabled, skipping registration",
			logger.String("service", info.Name()))
		return nil
	}

	// In a full implementation, this would register the service with the xDS control plane
	// For Istio, this is typically handled by the Kubernetes service registry
	logger.WarnField("xDS service registration not fully implemented",
		logger.String("service", info.Name()),
		logger.String("note", "Istio typically uses Kubernetes service discovery"))

	return nil
}

// Deregister deregisters a service instance
func (r *registry) Deregister(ctx context.Context, info registry2.Instance) error {
	if !r.config.EnableRegistration {
		return nil
	}

	logger.InfoField("xDS service deregistration",
		logger.String("service", info.Name()))

	return nil
}

// Name returns the registry name
func (r *registry) Name() string {
	return name
}
