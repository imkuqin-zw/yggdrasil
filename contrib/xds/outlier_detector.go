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
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

// OutlierDetectionConfig holds outlier detection configuration
type OutlierDetectionConfig struct {
	Consecutive5xx                 uint32
	ConsecutiveGatewayFailure      uint32
	ConsecutiveLocalOriginFailure  uint32
	Interval                       time.Duration
	BaseEjectionTime               time.Duration
	MaxEjectionTime                time.Duration
	MaxEjectionPercent             uint32
	EnforcingConsecutive5xx        uint32
	EnforcingSuccessRate           uint32
	SuccessRateMinimumHosts        uint32
	SuccessRateRequestVolume       uint32
	SuccessRateStdevFactor         uint32
	FailurePercentageThreshold     uint32
	EnforcingFailurePercentage     uint32
	FailurePercentageMinimumHosts  uint32
	FailurePercentageRequestVolume uint32
	SplitExternalLocalOriginErrors bool
}

// DefaultOutlierDetectionConfig returns default outlier detection configuration
func DefaultOutlierDetectionConfig() *OutlierDetectionConfig {
	return &OutlierDetectionConfig{
		Consecutive5xx:                 5,
		ConsecutiveGatewayFailure:      5,
		ConsecutiveLocalOriginFailure:  5,
		Interval:                       10 * time.Second,
		BaseEjectionTime:               30 * time.Second,
		MaxEjectionTime:                300 * time.Second,
		MaxEjectionPercent:             10,
		EnforcingConsecutive5xx:        100,
		EnforcingSuccessRate:           100,
		SuccessRateMinimumHosts:        5,
		SuccessRateRequestVolume:       100,
		SuccessRateStdevFactor:         1900, // 1.9 * 1000
		FailurePercentageThreshold:     85,
		EnforcingFailurePercentage:     0,
		FailurePercentageMinimumHosts:  5,
		FailurePercentageRequestVolume: 50,
		SplitExternalLocalOriginErrors: false,
	}
}

// EndpointStats tracks statistics for a single endpoint
type EndpointStats struct {
	address string

	// Request counters
	totalRequests   uint64
	successCount    uint64
	failureCount    uint64
	localFailures   uint64
	gatewayFailures uint64

	// Consecutive error tracking
	consecutive5xx            uint32
	consecutiveGatewayFailure uint32
	consecutiveLocalFailure   uint32

	// Ejection state
	ejected       bool
	ejectionTime  time.Time
	ejectionCount uint32

	mu sync.RWMutex
}

// OutlierDetector implements error-rate based outlier detection
type OutlierDetector struct {
	config    *OutlierDetectionConfig
	endpoints map[string]*EndpointStats
	mu        sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Statistics
	totalEjections uint64
}

// NewOutlierDetector creates a new outlier detector
func NewOutlierDetector(config *OutlierDetectionConfig) *OutlierDetector {
	if config == nil {
		config = DefaultOutlierDetectionConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &OutlierDetector{
		config:    config,
		endpoints: make(map[string]*EndpointStats),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins periodic health check sweeps
func (od *OutlierDetector) Start() {
	if od.config.Interval == 0 {
		return
	}

	od.wg.Add(1)
	go od.runHealthSweep()

	logger.InfoField("outlier detector started",
		logger.Duration("interval", od.config.Interval))
}

// Stop stops the outlier detector
func (od *OutlierDetector) Stop() {
	od.cancel()
	od.wg.Wait()
	logger.InfoField("outlier detector stopped")
}

// runHealthSweep runs periodic health check sweeps
func (od *OutlierDetector) runHealthSweep() {
	defer od.wg.Done()

	ticker := time.NewTicker(od.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-od.ctx.Done():
			return
		case <-ticker.C:
			od.performHealthSweep()
		}
	}
}

// performHealthSweep performs a health check sweep
func (od *OutlierDetector) performHealthSweep() {
	od.mu.RLock()
	endpoints := make([]*EndpointStats, 0, len(od.endpoints))
	for _, ep := range od.endpoints {
		endpoints = append(endpoints, ep)
	}
	od.mu.RUnlock()

	now := time.Now()

	// Check for recovery (uneject endpoints whose ejection time has expired)
	for _, ep := range endpoints {
		ep.mu.Lock()
		if ep.ejected && now.After(ep.ejectionTime) {
			ep.ejected = false
			logger.InfoField("endpoint recovered from ejection",
				logger.String("endpoint", ep.address),
				logger.Uint32("ejectionCount", ep.ejectionCount))
		}
		ep.mu.Unlock()
	}

	// Perform outlier detection
	od.detectSuccessRateOutliers(endpoints)
	od.detectFailurePercentageOutliers(endpoints)

	// Reset interval statistics
	for _, ep := range endpoints {
		ep.mu.Lock()
		ep.totalRequests = 0
		ep.successCount = 0
		ep.failureCount = 0
		ep.localFailures = 0
		ep.gatewayFailures = 0
		ep.mu.Unlock()
	}
}

// ReportResult reports request result for outlier detection
func (od *OutlierDetector) ReportResult(endpoint string, err error, statusCode int) {
	od.mu.Lock()
	ep, exists := od.endpoints[endpoint]
	if !exists {
		ep = &EndpointStats{
			address: endpoint,
		}
		od.endpoints[endpoint] = ep
	}
	od.mu.Unlock()

	ep.mu.Lock()
	defer ep.mu.Unlock()

	atomic.AddUint64(&ep.totalRequests, 1)

	isSuccess := err == nil && statusCode >= 200 && statusCode < 500
	is5xx := statusCode >= 500 && statusCode < 600
	isGatewayFailure := statusCode == 502 || statusCode == 503 || statusCode == 504
	isLocalFailure := err != nil

	if isSuccess {
		atomic.AddUint64(&ep.successCount, 1)
		// Reset consecutive error counters on success
		ep.consecutive5xx = 0
		ep.consecutiveGatewayFailure = 0
		ep.consecutiveLocalFailure = 0
	} else {
		atomic.AddUint64(&ep.failureCount, 1)

		// Track consecutive errors
		if is5xx {
			ep.consecutive5xx++
			od.checkConsecutive5xx(ep)
		}

		if isGatewayFailure {
			atomic.AddUint64(&ep.gatewayFailures, 1)
			ep.consecutiveGatewayFailure++
			od.checkConsecutiveGatewayFailure(ep)
		}

		if isLocalFailure {
			atomic.AddUint64(&ep.localFailures, 1)
			ep.consecutiveLocalFailure++
			od.checkConsecutiveLocalFailure(ep)
		}
	}
}

// checkConsecutive5xx checks for consecutive 5xx errors
func (od *OutlierDetector) checkConsecutive5xx(ep *EndpointStats) {
	if od.config.Consecutive5xx == 0 || od.config.EnforcingConsecutive5xx == 0 {
		return
	}

	if ep.consecutive5xx >= od.config.Consecutive5xx {
		if od.shouldEnforce(od.config.EnforcingConsecutive5xx) {
			od.ejectEndpoint(ep, "consecutive_5xx")
		}
	}
}

// checkConsecutiveGatewayFailure checks for consecutive gateway failures
func (od *OutlierDetector) checkConsecutiveGatewayFailure(ep *EndpointStats) {
	if od.config.ConsecutiveGatewayFailure == 0 {
		return
	}

	if ep.consecutiveGatewayFailure >= od.config.ConsecutiveGatewayFailure {
		if od.shouldEnforce(od.config.EnforcingConsecutive5xx) {
			od.ejectEndpoint(ep, "consecutive_gateway_failure")
		}
	}
}

// checkConsecutiveLocalFailure checks for consecutive local failures
func (od *OutlierDetector) checkConsecutiveLocalFailure(ep *EndpointStats) {
	if od.config.ConsecutiveLocalOriginFailure == 0 {
		return
	}

	if ep.consecutiveLocalFailure >= od.config.ConsecutiveLocalOriginFailure {
		if od.shouldEnforce(od.config.EnforcingConsecutive5xx) {
			od.ejectEndpoint(ep, "consecutive_local_failure")
		}
	}
}

// detectSuccessRateOutliers detects outliers based on success rate
func (od *OutlierDetector) detectSuccessRateOutliers(endpoints []*EndpointStats) {
	if od.config.EnforcingSuccessRate == 0 {
		return
	}

	// Filter endpoints with sufficient request volume
	var validEndpoints []*EndpointStats
	for _, ep := range endpoints {
		ep.mu.RLock()
		total := atomic.LoadUint64(&ep.totalRequests)
		ep.mu.RUnlock()

		if total >= uint64(od.config.SuccessRateRequestVolume) {
			validEndpoints = append(validEndpoints, ep)
		}
	}

	if len(validEndpoints) < int(od.config.SuccessRateMinimumHosts) {
		return
	}

	// Calculate success rates
	successRates := make([]float64, len(validEndpoints))
	var sum float64
	for i, ep := range validEndpoints {
		ep.mu.RLock()
		total := atomic.LoadUint64(&ep.totalRequests)
		success := atomic.LoadUint64(&ep.successCount)
		ep.mu.RUnlock()

		if total > 0 {
			successRates[i] = float64(success) / float64(total) * 100
			sum += successRates[i]
		}
	}

	// Calculate mean and standard deviation
	mean := sum / float64(len(validEndpoints))
	var variance float64
	for _, rate := range successRates {
		diff := rate - mean
		variance += diff * diff
	}
	stddev := math.Sqrt(variance / float64(len(validEndpoints)))

	// Eject outliers
	threshold := mean - (float64(od.config.SuccessRateStdevFactor) / 1000.0 * stddev)
	for i, ep := range validEndpoints {
		if successRates[i] < threshold {
			if od.shouldEnforce(od.config.EnforcingSuccessRate) {
				ep.mu.Lock()
				od.ejectEndpoint(ep, "success_rate")
				ep.mu.Unlock()
			}
		}
	}
}

// detectFailurePercentageOutliers detects outliers based on failure percentage
func (od *OutlierDetector) detectFailurePercentageOutliers(endpoints []*EndpointStats) {
	if od.config.EnforcingFailurePercentage == 0 || od.config.FailurePercentageThreshold == 0 {
		return
	}

	var validCount int
	for _, ep := range endpoints {
		ep.mu.RLock()
		total := atomic.LoadUint64(&ep.totalRequests)
		ep.mu.RUnlock()

		if total >= uint64(od.config.FailurePercentageRequestVolume) {
			validCount++
		}
	}

	if validCount < int(od.config.FailurePercentageMinimumHosts) {
		return
	}

	for _, ep := range endpoints {
		ep.mu.Lock()
		total := atomic.LoadUint64(&ep.totalRequests)
		failures := atomic.LoadUint64(&ep.failureCount)

		if total >= uint64(od.config.FailurePercentageRequestVolume) {
			failurePercentage := float64(failures) / float64(total) * 100
			if failurePercentage >= float64(od.config.FailurePercentageThreshold) {
				if od.shouldEnforce(od.config.EnforcingFailurePercentage) {
					od.ejectEndpoint(ep, "failure_percentage")
				}
			}
		}
		ep.mu.Unlock()
	}
}

// ejectEndpoint ejects an endpoint
func (od *OutlierDetector) ejectEndpoint(ep *EndpointStats, reason string) {
	if ep.ejected {
		return
	}

	// Check max ejection percentage
	od.mu.RLock()
	totalEndpoints := len(od.endpoints)
	od.mu.RUnlock()

	var ejectedCount int
	od.mu.RLock()
	for _, e := range od.endpoints {
		e.mu.RLock()
		if e.ejected {
			ejectedCount++
		}
		e.mu.RUnlock()
	}
	od.mu.RUnlock()

	maxEjected := int(float64(totalEndpoints) * float64(od.config.MaxEjectionPercent) / 100.0)
	if ejectedCount >= maxEjected && maxEjected > 0 {
		logger.DebugField("max ejection percentage reached, not ejecting",
			logger.String("endpoint", ep.address),
			logger.Int("ejected", ejectedCount),
			logger.Int("total", totalEndpoints))
		return
	}

	ep.ejected = true
	ep.ejectionCount++

	// Calculate ejection time with exponential backoff
	ejectionDuration := od.config.BaseEjectionTime * time.Duration(ep.ejectionCount)
	if ejectionDuration > od.config.MaxEjectionTime {
		ejectionDuration = od.config.MaxEjectionTime
	}
	ep.ejectionTime = time.Now().Add(ejectionDuration)

	atomic.AddUint64(&od.totalEjections, 1)

	logger.WarnField("endpoint ejected",
		logger.String("endpoint", ep.address),
		logger.String("reason", reason),
		logger.Uint32("ejectionCount", ep.ejectionCount),
		logger.Duration("ejectionDuration", ejectionDuration))
}

// shouldEnforce determines if enforcement should happen based on percentage
func (od *OutlierDetector) shouldEnforce(enforcingPercentage uint32) bool {
	if enforcingPercentage == 0 {
		return false
	}
	if enforcingPercentage >= 100 {
		return true
	}
	// For simplicity, always enforce if percentage > 0
	// In production, this should use random sampling
	return true
}

// IsEjected checks if an endpoint is currently ejected
func (od *OutlierDetector) IsEjected(endpoint string) bool {
	od.mu.RLock()
	ep, exists := od.endpoints[endpoint]
	od.mu.RUnlock()

	if !exists {
		return false
	}

	ep.mu.RLock()
	defer ep.mu.RUnlock()

	return ep.ejected
}

// GetStats returns outlier detection statistics
func (od *OutlierDetector) GetStats() map[string]interface{} {
	od.mu.RLock()
	defer od.mu.RUnlock()

	var ejectedCount int
	for _, ep := range od.endpoints {
		ep.mu.RLock()
		if ep.ejected {
			ejectedCount++
		}
		ep.mu.RUnlock()
	}

	return map[string]interface{}{
		"total_endpoints": len(od.endpoints),
		"ejected_count":   ejectedCount,
		"total_ejections": atomic.LoadUint64(&od.totalEjections),
	}
}
