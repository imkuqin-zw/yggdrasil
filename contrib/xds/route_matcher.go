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
	"math/rand"
	"regexp"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

// Request represents an incoming request for routing
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Host    string
}

// RouteMatch represents a matched route result
type RouteMatch struct {
	Cluster          string
	WeightedClusters *route.WeightedCluster
	ClusterHeader    string
}

// RouteMatcher implements route matching logic
type RouteMatcher struct {
	resourceManager *ResourceManager
}

// NewRouteMatcher creates a new route matcher
func NewRouteMatcher(rm *ResourceManager) *RouteMatcher {
	return &RouteMatcher{
		resourceManager: rm,
	}
}

// Match matches a request against route configuration
func (rm *RouteMatcher) Match(routeConfigName string, req *Request) (*RouteMatch, error) {
	// Get route configuration
	routeConfig, err := rm.resourceManager.GetRoute(routeConfigName)
	if err != nil {
		return nil, err
	}

	// Find matching virtual host
	vh := rm.findVirtualHost(routeConfig, req.Host)
	if vh == nil {
		return nil, fmt.Errorf("no matching virtual host for: %s", req.Host)
	}

	// Find matching route
	for _, r := range vh.Routes {
		if rm.matchRoute(r, req) {
			return rm.extractRouteAction(r)
		}
	}

	return nil, fmt.Errorf("no matching route for path: %s", req.Path)
}

// findVirtualHost finds a virtual host that matches the request host
func (rm *RouteMatcher) findVirtualHost(config *route.RouteConfiguration, host string) *route.VirtualHost {
	for _, vh := range config.VirtualHosts {
		for _, domain := range vh.Domains {
			if rm.matchDomain(domain, host) {
				return vh
			}
		}
	}
	return nil
}

// matchDomain checks if a domain pattern matches the host
func (rm *RouteMatcher) matchDomain(pattern, host string) bool {
	// Exact match
	if pattern == host {
		return true
	}

	// Wildcard match
	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(host, suffix)
	}

	// Catch-all
	if pattern == "*" {
		return true
	}

	return false
}

// matchRoute checks if a route matches the request
func (rm *RouteMatcher) matchRoute(r *route.Route, req *Request) bool {
	// Match path
	if !rm.matchPath(r.Match, req.Path) {
		return false
	}

	// Match headers
	if !rm.matchHeaders(r.Match, req.Headers) {
		return false
	}

	// Match query parameters (if needed)
	// TODO: implement query parameter matching

	return true
}

// matchPath checks if the path matches
func (rm *RouteMatcher) matchPath(match *route.RouteMatch, path string) bool {
	switch pathSpec := match.PathSpecifier.(type) {
	case *route.RouteMatch_Prefix:
		return strings.HasPrefix(path, pathSpec.Prefix)

	case *route.RouteMatch_Path:
		return path == pathSpec.Path

	case *route.RouteMatch_SafeRegex:
		re, err := regexp.Compile(pathSpec.SafeRegex.Regex)
		if err != nil {
			return false
		}
		return re.MatchString(path)

	default:
		return false
	}
}

// matchHeaders checks if headers match
func (rm *RouteMatcher) matchHeaders(match *route.RouteMatch, headers map[string]string) bool {
	for _, headerMatcher := range match.Headers {
		if !rm.matchHeader(headerMatcher, headers) {
			return false
		}
	}
	return true
}

// matchHeader checks if a single header matches
func (rm *RouteMatcher) matchHeader(matcher *route.HeaderMatcher, headers map[string]string) bool {
	headerValue, ok := headers[matcher.Name]
	if !ok {
		// Header not present
		return matcher.InvertMatch
	}

	matched := false

	switch headerMatchSpec := matcher.HeaderMatchSpecifier.(type) {
	case *route.HeaderMatcher_ExactMatch:
		matched = headerValue == headerMatchSpec.ExactMatch

	case *route.HeaderMatcher_PrefixMatch:
		matched = strings.HasPrefix(headerValue, headerMatchSpec.PrefixMatch)

	case *route.HeaderMatcher_SuffixMatch:
		matched = strings.HasSuffix(headerValue, headerMatchSpec.SuffixMatch)

	case *route.HeaderMatcher_SafeRegexMatch:
		re, err := regexp.Compile(headerMatchSpec.SafeRegexMatch.Regex)
		if err == nil {
			matched = re.MatchString(headerValue)
		}

	case *route.HeaderMatcher_PresentMatch:
		matched = headerMatchSpec.PresentMatch

	case *route.HeaderMatcher_ContainsMatch:
		matched = strings.Contains(headerValue, headerMatchSpec.ContainsMatch)

	case *route.HeaderMatcher_StringMatch:
		matched = rm.matchStringMatcher(headerMatchSpec.StringMatch, headerValue)
	}

	if matcher.InvertMatch {
		return !matched
	}
	return matched
}

// matchStringMatcher matches against a StringMatcher
func (rm *RouteMatcher) matchStringMatcher(matcher *matcherv3.StringMatcher, value string) bool {
	switch m := matcher.MatchPattern.(type) {
	case *matcherv3.StringMatcher_Exact:
		return value == m.Exact

	case *matcherv3.StringMatcher_Prefix:
		return strings.HasPrefix(value, m.Prefix)

	case *matcherv3.StringMatcher_Suffix:
		return strings.HasSuffix(value, m.Suffix)

	case *matcherv3.StringMatcher_SafeRegex:
		re, err := regexp.Compile(m.SafeRegex.Regex)
		if err == nil {
			return re.MatchString(value)
		}

	case *matcherv3.StringMatcher_Contains:
		return strings.Contains(value, m.Contains)
	}

	return false
}

// extractRouteAction extracts the route action from a matched route
func (rm *RouteMatcher) extractRouteAction(r *route.Route) (*RouteMatch, error) {
	routeAction := r.GetRoute()
	if routeAction == nil {
		return nil, fmt.Errorf("route has no action")
	}

	match := &RouteMatch{}

	switch clusterSpec := routeAction.ClusterSpecifier.(type) {
	case *route.RouteAction_Cluster:
		match.Cluster = clusterSpec.Cluster

	case *route.RouteAction_WeightedClusters:
		match.WeightedClusters = clusterSpec.WeightedClusters

	case *route.RouteAction_ClusterHeader:
		match.ClusterHeader = clusterSpec.ClusterHeader

	default:
		return nil, fmt.Errorf("unsupported cluster specifier")
	}

	return match, nil
}

// SelectWeightedCluster selects a cluster from weighted clusters based on weights
func SelectWeightedCluster(wc *route.WeightedCluster) string {
	if wc == nil || len(wc.Clusters) == 0 {
		return ""
	}

	// Calculate total weight
	var totalWeight uint32
	for _, c := range wc.Clusters {
		weight := c.Weight.GetValue()
		if weight == 0 {
			weight = 1 // Default weight
		}
		totalWeight += weight
	}

	// Random selection based on weight
	r := rand.Uint32() % totalWeight
	var sum uint32

	for _, c := range wc.Clusters {
		weight := c.Weight.GetValue()
		if weight == 0 {
			weight = 1
		}
		sum += weight
		if r < sum {
			return c.Name
		}
	}

	// Fallback to first cluster
	return wc.Clusters[0].Name
}
