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

	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"google.golang.org/protobuf/types/known/anypb"
)

// LDSHandler handles Listener Discovery Service updates
type LDSHandler struct {
	resourceManager *ResourceManager
	client          *Client
}

// NewLDSHandler creates a new LDS handler
func NewLDSHandler(rm *ResourceManager, client *Client) *LDSHandler {
	return &LDSHandler{
		resourceManager: rm,
		client:          client,
	}
}

// HandleUpdate processes LDS updates
func (h *LDSHandler) HandleUpdate(resources []interface{}) error {
	listeners := make([]*listener.Listener, 0, len(resources))

	for _, res := range resources {
		anyRes, ok := res.(*anypb.Any)
		if !ok {
			logger.WarnField("unexpected resource type in LDS update")
			continue
		}

		// Unmarshal Listener
		lis := &listener.Listener{}
		if err := anyRes.UnmarshalTo(lis); err != nil {
			logger.ErrorField("failed to unmarshal Listener", logger.Err(err))
			continue
		}

		listeners = append(listeners, lis)

		// Extract HTTP Connection Manager and route config
		if err := h.processListener(lis); err != nil {
			logger.ErrorField("failed to process listener",
				logger.String("listener", lis.Name),
				logger.Err(err))
		}
	}

	// Update resource manager
	return h.resourceManager.UpdateLDS(listeners)
}

// processListener extracts HTTP Connection Manager and handles route configuration
func (h *LDSHandler) processListener(lis *listener.Listener) error {
	// Find HTTP Connection Manager filter
	httpConnMgr, err := h.extractHTTPConnectionManager(lis)
	if err != nil {
		return err
	}

	// Handle route configuration
	switch rds := httpConnMgr.RouteSpecifier.(type) {
	case *hcm.HttpConnectionManager_Rds:
		// RDS reference - subscribe to RDS
		routeConfigName := rds.Rds.RouteConfigName
		h.resourceManager.SetListenerRouteMapping(lis.Name, routeConfigName)

		logger.InfoField("listener references RDS",
			logger.String("listener", lis.Name),
			logger.String("route", routeConfigName))

		// Subscribe to RDS if not already subscribed
		// This will be handled by the client's subscription management

	case *hcm.HttpConnectionManager_RouteConfig:
		// Inline route configuration
		routeConfig := rds.RouteConfig
		h.resourceManager.SetListenerRouteMapping(lis.Name, routeConfig.Name)

		logger.InfoField("listener has inline route config",
			logger.String("listener", lis.Name),
			logger.String("route", routeConfig.Name))

		// Update resource manager with inline route
		if err := h.resourceManager.UpdateRDS([]*route.RouteConfiguration{routeConfig}); err != nil {
			return fmt.Errorf("failed to update inline route: %w", err)
		}

	default:
		return fmt.Errorf("unsupported route specifier type")
	}

	return nil
}

// extractHTTPConnectionManager extracts the HTTP Connection Manager from a listener
func (h *LDSHandler) extractHTTPConnectionManager(lis *listener.Listener) (*hcm.HttpConnectionManager, error) {
	for _, filterChain := range lis.FilterChains {
		for _, filter := range filterChain.Filters {
			if filter.Name == "envoy.filters.network.http_connection_manager" {
				httpConnMgr := &hcm.HttpConnectionManager{}

				switch c := filter.ConfigType.(type) {
				case *listener.Filter_TypedConfig:
					if err := c.TypedConfig.UnmarshalTo(httpConnMgr); err != nil {
						return nil, fmt.Errorf("failed to unmarshal HTTP connection manager: %w", err)
					}
					return httpConnMgr, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("HTTP connection manager not found in listener")
}
