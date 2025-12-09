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

package snapshot

import (
	"fmt"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Builder builds xDS snapshots from configuration
type Builder struct {
	version string
}

// NewBuilder creates a new snapshot builder
func NewBuilder(version string) *Builder {
	return &Builder{
		version: version,
	}
}

// BuildSnapshot builds a complete xDS snapshot from configuration
func (b *Builder) BuildSnapshot(config *XDSConfig) (*cache.Snapshot, error) {
	// Build resources
	clusters := b.buildClusters(config.Clusters)
	endpoints := b.buildEndpoints(config.Endpoints)
	listeners := b.buildListeners(config.Listeners)
	routes := b.buildRoutes(config.Routes)
	// Create snapshot
	snapshot, err := cache.NewSnapshot(
		b.version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  clusters,
			resource.EndpointType: endpoints,
			resource.ListenerType: listeners,
			resource.RouteType:    routes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return snapshot, nil
}

// buildClusters builds cluster resources
func (b *Builder) buildClusters(configs []Cluster) []types.Resource {
	var clusters []types.Resource

	for _, cfg := range configs {
		c := &cluster.Cluster{
			Name:                 cfg.Name,
			ConnectTimeout:       durationpb.New(ParseDuration(cfg.ConnectTimeout, 5*time.Second)),
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
			LbPolicy:             b.parseLbPolicy(cfg.LbPolicy),
			EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
				ServiceName: cfg.Name,
				EdsConfig: &core.ConfigSource{
					ResourceApiVersion: core.ApiVersion_V3,
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
			},
		}
		clusters = append(clusters, c)
	}

	return clusters
}

// buildEndpoints builds endpoint resources (ClusterLoadAssignment)
func (b *Builder) buildEndpoints(configs []Endpoint) []types.Resource {
	var endpoints []types.Resource

	for _, cfg := range configs {
		var lbEndpoints []*endpoint.LbEndpoint

		for _, ep := range cfg.Endpoints {
			lbEndpoints = append(lbEndpoints, &endpoint.LbEndpoint{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  ep.Address,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: ep.Port,
									},
								},
							},
						},
					},
				},
			})
		}

		cla := &endpoint.ClusterLoadAssignment{
			ClusterName: cfg.ClusterName,
			Endpoints: []*endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: lbEndpoints,
				},
			},
		}
		endpoints = append(endpoints, cla)
	}

	return endpoints
}

// buildListeners builds listener resources
func (b *Builder) buildListeners(configs []Listener) []types.Resource {
	var listeners []types.Resource

	for _, cfg := range configs {
		// For simplicity, we'll create an HTTP connection manager
		manager := &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "ingress_http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource: &core.ConfigSource{
						ResourceApiVersion: core.ApiVersion_V3,
						ConfigSourceSpecifier: &core.ConfigSource_Ads{
							Ads: &core.AggregatedConfigSource{},
						},
					},
					RouteConfigName: b.getRouteConfigName(cfg),
				},
			},
			HttpFilters: []*hcm.HttpFilter{
				{
					Name: "envoy.filters.http.router",
				},
			},
		}

		pbst, err := anypb.New(manager)
		if err != nil {
			continue
		}

		l := &listener.Listener{
			Name: cfg.Name,
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Protocol: core.SocketAddress_TCP,
						Address:  cfg.Address,
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: cfg.Port,
						},
					},
				},
			},
			FilterChains: []*listener.FilterChain{
				{
					Filters: []*listener.Filter{
						{
							Name: "envoy.filters.network.http_connection_manager",
							ConfigType: &listener.Filter_TypedConfig{
								TypedConfig: pbst,
							},
						},
					},
				},
			},
		}
		listeners = append(listeners, l)
	}

	return listeners
}

// buildRoutes builds route configuration resources
func (b *Builder) buildRoutes(configs []Route) []types.Resource {
	var routes []types.Resource

	for _, cfg := range configs {
		var virtualHosts []*route.VirtualHost

		for _, vh := range cfg.VirtualHosts {
			var routeMatches []*route.Route
			for _, rm := range vh.Routes {
				var match = &route.RouteMatch{}

				if rm.Match.Path != nil {
					if rm.Match.Path.Prefix != "" {
						match.PathSpecifier = &route.RouteMatch_Prefix{
							Prefix: rm.Match.Path.Prefix,
						}
					} else if rm.Match.Path.Path != "" {
						match.PathSpecifier = &route.RouteMatch_Path{
							Path: rm.Match.Path.Path,
						}
					}
				}

				var headers []*route.HeaderMatcher
				for _, item := range rm.Match.Headers {
					stringMatch := &matcher.StringMatcher{}
					switch item.Pattern {
					case "exact":
						stringMatch.MatchPattern = &matcher.StringMatcher_Exact{
							Exact: item.Value,
						}
					case "prefix":
						stringMatch.MatchPattern = &matcher.StringMatcher_Prefix{
							Prefix: item.Value,
						}
					case "suffix":
						stringMatch.MatchPattern = &matcher.StringMatcher_Suffix{
							Suffix: item.Value,
						}
					case "safeRegex":
						stringMatch.MatchPattern = &matcher.StringMatcher_SafeRegex{
							SafeRegex: &matcher.RegexMatcher{
								Regex: item.Value,
							},
						}
					case "contains":
						stringMatch.MatchPattern = &matcher.StringMatcher_Contains{
							Contains: item.Value,
						}
					}
					headers = append(headers, &route.HeaderMatcher{
						Name: item.Name,
						HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
							StringMatch: stringMatch,
						},
					})
				}
				match.Headers = headers

				routeMatches = append(routeMatches, &route.Route{
					Match: match,
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: rm.Route.Cluster,
							},
						},
					},
				})
			}

			virtualHosts = append(virtualHosts, &route.VirtualHost{
				Name:    vh.Name,
				Domains: vh.Domains,
				Routes:  routeMatches,
			})
		}

		rc := &route.RouteConfiguration{
			Name:         cfg.Name,
			VirtualHosts: virtualHosts,
		}
		routes = append(routes, rc)
	}

	return routes
}

// Helper functions

func (b *Builder) parseClusterType(t string) cluster.Cluster_DiscoveryType {
	switch t {
	case "STATIC":
		return cluster.Cluster_STATIC
	case "STRICT_DNS":
		return cluster.Cluster_STRICT_DNS
	case "LOGICAL_DNS":
		return cluster.Cluster_LOGICAL_DNS
	case "EDS":
		return cluster.Cluster_EDS
	default:
		return cluster.Cluster_STATIC
	}
}

func (b *Builder) parseLbPolicy(policy string) cluster.Cluster_LbPolicy {
	switch policy {
	case "ROUND_ROBIN":
		return cluster.Cluster_ROUND_ROBIN
	case "LEAST_REQUEST":
		return cluster.Cluster_LEAST_REQUEST
	case "RING_HASH":
		return cluster.Cluster_RING_HASH
	case "RANDOM":
		return cluster.Cluster_RANDOM
	case "MAGLEV":
		return cluster.Cluster_MAGLEV
	default:
		return cluster.Cluster_ROUND_ROBIN
	}
}

func (b *Builder) makeEndpoint(clusterName string) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
	}
}

func (b *Builder) getRouteConfigName(l Listener) string {
	// Extract route config name from filter chains
	for _, fc := range l.FilterChains {
		for _, f := range fc.Filters {
			if f.RouteConfigName != "" {
				return f.RouteConfigName
			}
		}
	}
	return "local_route"
}
