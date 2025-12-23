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
	"errors"
	"fmt"
)

// Error types for xDS operations
var (
	// ErrClientClosed is returned when the client is closed
	ErrClientClosed = errors.New("xDS client is closed")

	// ErrConnectionFailed is returned when connection to xDS server fails
	ErrConnectionFailed = errors.New("failed to connect to xDS server")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid xDS configuration")

	// ErrNoRouteMatch is returned when no route matches the request
	ErrNoRouteMatch = errors.New("no route match found")

	// ErrNoCluster is returned when no cluster is found
	ErrNoCluster = errors.New("no cluster found")

	// ErrNoEndpoints is returned when no endpoints are available
	ErrNoEndpoints = errors.New("no endpoints available")

	// ErrCircuitBreakerOpen is returned when circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker open")

	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrEndpointEjected is returned when endpoint is ejected by outlier detection
	ErrEndpointEjected = errors.New("endpoint ejected by outlier detection")
)

// ErrSubscriptionFailed creates a subscription error
func ErrSubscriptionFailed(resourceType, resourceName string, err error) error {
	return fmt.Errorf("failed to subscribe to %s/%s: %w", resourceType, resourceName, err)
}

// ErrUpdateFailed creates an update error
func ErrUpdateFailed(resourceType string, err error) error {
	return fmt.Errorf("xds: failed to process %s update: %w", resourceType, err)
}
