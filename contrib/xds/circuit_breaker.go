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
	"sync/atomic"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxConnections     uint32
	MaxPendingRequests uint32
	MaxRequests        uint32
	MaxRetries         uint32
}

// CircuitBreaker implements resource-based circuit breaking
type CircuitBreaker struct {
	config *CircuitBreakerConfig

	// Atomic counters for thread-safe tracking
	activeConnections uint32
	pendingRequests   uint32
	activeRequests    uint32
	activeRetries     uint32
	rejectedRequests  uint64
	rejectedRetries   uint64
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			MaxConnections:     1024,
			MaxPendingRequests: 1024,
			MaxRequests:        1024,
			MaxRetries:         3,
		}
	}

	return &CircuitBreaker{
		config: config,
	}
}

// ResourceType represents the type of resource to acquire
type ResourceType int

const (
	ResourceConnection ResourceType = iota
	ResourcePendingRequest
	ResourceRequest
	ResourceRetry
)

// TryAcquire attempts to acquire a resource slot
func (cb *CircuitBreaker) TryAcquire(resourceType ResourceType) bool {
	switch resourceType {
	case ResourceConnection:
		return cb.tryAcquireConnection()
	case ResourcePendingRequest:
		return cb.tryAcquirePendingRequest()
	case ResourceRequest:
		return cb.tryAcquireRequest()
	case ResourceRetry:
		return cb.tryAcquireRetry()
	default:
		return false
	}
}

// Release releases a resource slot
func (cb *CircuitBreaker) Release(resourceType ResourceType) {
	switch resourceType {
	case ResourceConnection:
		cb.releaseConnection()
	case ResourcePendingRequest:
		cb.releasePendingRequest()
	case ResourceRequest:
		cb.releaseRequest()
	case ResourceRetry:
		cb.releaseRetry()
	}
}

// tryAcquireConnection attempts to acquire a connection slot
func (cb *CircuitBreaker) tryAcquireConnection() bool {
	if cb.config.MaxConnections == 0 {
		return true // No limit
	}

	for {
		current := atomic.LoadUint32(&cb.activeConnections)
		if current >= cb.config.MaxConnections {
			logger.DebugField("circuit breaker: max connections reached",
				logger.Uint32("current", current),
				logger.Uint32("max", cb.config.MaxConnections))
			return false
		}

		if atomic.CompareAndSwapUint32(&cb.activeConnections, current, current+1) {
			return true
		}
	}
}

// releaseConnection releases a connection slot
func (cb *CircuitBreaker) releaseConnection() {
	atomic.AddUint32(&cb.activeConnections, ^uint32(0)) // Decrement
}

// tryAcquirePendingRequest attempts to acquire a pending request slot
func (cb *CircuitBreaker) tryAcquirePendingRequest() bool {
	if cb.config.MaxPendingRequests == 0 {
		return true // No limit
	}

	for {
		current := atomic.LoadUint32(&cb.pendingRequests)
		if current >= cb.config.MaxPendingRequests {
			logger.DebugField("circuit breaker: max pending requests reached",
				logger.Uint32("current", current),
				logger.Uint32("max", cb.config.MaxPendingRequests))
			return false
		}

		if atomic.CompareAndSwapUint32(&cb.pendingRequests, current, current+1) {
			return true
		}
	}
}

// releasePendingRequest releases a pending request slot
func (cb *CircuitBreaker) releasePendingRequest() {
	atomic.AddUint32(&cb.pendingRequests, ^uint32(0)) // Decrement
}

// tryAcquireRequest attempts to acquire a request slot
func (cb *CircuitBreaker) tryAcquireRequest() bool {
	if cb.config.MaxRequests == 0 {
		return true // No limit
	}

	for {
		current := atomic.LoadUint32(&cb.activeRequests)
		if current >= cb.config.MaxRequests {
			atomic.AddUint64(&cb.rejectedRequests, 1)
			logger.DebugField("circuit breaker: max requests reached",
				logger.Uint32("current", current),
				logger.Uint32("max", cb.config.MaxRequests))
			return false
		}

		if atomic.CompareAndSwapUint32(&cb.activeRequests, current, current+1) {
			return true
		}
	}
}

// releaseRequest releases a request slot
func (cb *CircuitBreaker) releaseRequest() {
	atomic.AddUint32(&cb.activeRequests, ^uint32(0)) // Decrement
}

// tryAcquireRetry attempts to acquire a retry slot
func (cb *CircuitBreaker) tryAcquireRetry() bool {
	if cb.config.MaxRetries == 0 {
		return true // No limit
	}

	for {
		current := atomic.LoadUint32(&cb.activeRetries)
		if current >= cb.config.MaxRetries {
			atomic.AddUint64(&cb.rejectedRetries, 1)
			logger.DebugField("circuit breaker: max retries reached",
				logger.Uint32("current", current),
				logger.Uint32("max", cb.config.MaxRetries))
			return false
		}

		if atomic.CompareAndSwapUint32(&cb.activeRetries, current, current+1) {
			return true
		}
	}
}

// releaseRetry releases a retry slot
func (cb *CircuitBreaker) releaseRetry() {
	atomic.AddUint32(&cb.activeRetries, ^uint32(0)) // Decrement
}

// Stats represents circuit breaker statistics
type CircuitBreakerStats struct {
	ActiveConnections uint32
	PendingRequests   uint32
	ActiveRequests    uint32
	ActiveRetries     uint32
	RejectedRequests  uint64
	RejectedRetries   uint64
}

// GetStats returns current circuit breaker statistics
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	return CircuitBreakerStats{
		ActiveConnections: atomic.LoadUint32(&cb.activeConnections),
		PendingRequests:   atomic.LoadUint32(&cb.pendingRequests),
		ActiveRequests:    atomic.LoadUint32(&cb.activeRequests),
		ActiveRetries:     atomic.LoadUint32(&cb.activeRetries),
		RejectedRequests:  atomic.LoadUint64(&cb.rejectedRequests),
		RejectedRetries:   atomic.LoadUint64(&cb.rejectedRetries),
	}
}
