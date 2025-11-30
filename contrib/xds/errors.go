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
)

// Error types for xDS operations
var (
	ErrClientClosed     = fmt.Errorf("xds: client is closed")
	ErrConnectionFailed = fmt.Errorf("xds: connection failed")
	ErrInvalidResponse  = fmt.Errorf("xds: invalid response from server")
	ErrResourceNotFound = fmt.Errorf("xds: resource not found")
	ErrTimeout          = fmt.Errorf("xds: operation timeout")
	ErrNoEndpoints      = fmt.Errorf("xds: no available endpoints")
)

// ErrInvalidConfig creates a configuration error
func ErrInvalidConfig(msg string) error {
	return fmt.Errorf("xds: invalid config: %s", msg)
}

// ErrSubscriptionFailed creates a subscription error
func ErrSubscriptionFailed(resourceType, name string, err error) error {
	return fmt.Errorf("xds: failed to subscribe to %s resource '%s': %w", resourceType, name, err)
}

// ErrUpdateFailed creates an update error
func ErrUpdateFailed(resourceType string, err error) error {
	return fmt.Errorf("xds: failed to process %s update: %w", resourceType, err)
}
