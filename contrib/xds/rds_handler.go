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
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"google.golang.org/protobuf/types/known/anypb"
)

// RDSHandler handles Route Discovery Service updates
type RDSHandler struct {
	resourceManager *ResourceManager
}

// NewRDSHandler creates a new RDS handler
func NewRDSHandler(rm *ResourceManager) *RDSHandler {
	return &RDSHandler{
		resourceManager: rm,
	}
}

// HandleUpdate processes RDS updates
func (h *RDSHandler) HandleUpdate(resources []interface{}) error {
	routes := make([]*route.RouteConfiguration, 0, len(resources))

	for _, res := range resources {
		anyRes, ok := res.(*anypb.Any)
		if !ok {
			logger.WarnField("unexpected resource type in RDS update")
			continue
		}

		// Unmarshal RouteConfiguration
		routeConfig := &route.RouteConfiguration{}
		if err := anyRes.UnmarshalTo(routeConfig); err != nil {
			logger.ErrorField("failed to unmarshal RouteConfiguration", logger.Err(err))
			continue
		}

		routes = append(routes, routeConfig)

		logger.InfoField("RDS handler received route configuration",
			logger.String("name", routeConfig.Name),
			logger.Int("virtual_hosts", len(routeConfig.VirtualHosts)))
	}

	// Update resource manager
	return h.resourceManager.UpdateRDS(routes)
}
