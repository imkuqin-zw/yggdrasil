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
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxConnections:     10,
		MaxPendingRequests: 5,
		MaxRequests:        20,
		MaxRetries:         3,
	}

	cb := NewCircuitBreaker(config)

	// Test request acquisition
	t.Run("AcquireRequest", func(t *testing.T) {
		// Should be able to acquire up to MaxRequests
		for i := 0; i < int(config.MaxRequests); i++ {
			if !cb.TryAcquire(ResourceRequest) {
				t.Fatalf("Failed to acquire request %d/%d", i+1, config.MaxRequests)
			}
		}

		// Next acquisition should fail
		if cb.TryAcquire(ResourceRequest) {
			t.Fatal("Should not be able to acquire beyond MaxRequests")
		}

		// Release one and try again
		cb.Release(ResourceRequest)
		if !cb.TryAcquire(ResourceRequest) {
			t.Fatal("Should be able to acquire after release")
		}
	})

	// Test stats
	t.Run("Stats", func(t *testing.T) {
		stats := cb.GetStats()
		if stats.ActiveRequests != config.MaxRequests {
			t.Errorf("Expected %d active requests, got %d", config.MaxRequests, stats.ActiveRequests)
		}
	})
}

func TestOutlierDetector(t *testing.T) {
	config := &OutlierDetectionConfig{
		Consecutive5xx:            3,
		ConsecutiveGatewayFailure: 3,
		Interval:                  100 * time.Millisecond,
		BaseEjectionTime:          500 * time.Millisecond,
		MaxEjectionTime:           5 * time.Second,
		MaxEjectionPercent:        50,
		EnforcingConsecutive5xx:   100,
	}

	od := NewOutlierDetector(config)
	od.Start()
	defer od.Stop()

	endpoint := "192.168.1.1:8080"

	// Test consecutive 5xx detection
	t.Run("Consecutive5xx", func(t *testing.T) {
		// Report consecutive failures
		for i := 0; i < int(config.Consecutive5xx); i++ {
			od.ReportResult(endpoint, nil, 500)
		}

		// Endpoint should be ejected
		if !od.IsEjected(endpoint) {
			t.Error("Endpoint should be ejected after consecutive 5xx errors")
		}
	})

	// Test recovery
	t.Run("Recovery", func(t *testing.T) {
		// Wait for ejection time to expire
		time.Sleep(config.BaseEjectionTime + 200*time.Millisecond)

		// Endpoint should be recovered
		if od.IsEjected(endpoint) {
			t.Error("Endpoint should be recovered after ejection time")
		}
	})
}

func TestRateLimiter(t *testing.T) {
	config := &RateLimitConfig{
		MaxTokens:     10,
		TokensPerFill: 5,
		FillInterval:  100 * time.Millisecond,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	// Test token consumption
	t.Run("TokenConsumption", func(t *testing.T) {
		// Should be able to consume up to MaxTokens
		for i := 0; i < int(config.MaxTokens); i++ {
			if !rl.Allow() {
				t.Fatalf("Failed to acquire token %d/%d", i+1, config.MaxTokens)
			}
		}

		// Next acquisition should fail
		if rl.Allow() {
			t.Error("Should not be able to acquire beyond MaxTokens")
		}
	})

	// Test token refill
	t.Run("TokenRefill", func(t *testing.T) {
		// Wait for refill
		time.Sleep(config.FillInterval + 50*time.Millisecond)

		// Should be able to acquire tokens again
		acquired := 0
		for i := 0; i < int(config.TokensPerFill); i++ {
			if rl.Allow() {
				acquired++
			}
		}

		if acquired == 0 {
			t.Error("Should be able to acquire tokens after refill")
		}
	})
}
