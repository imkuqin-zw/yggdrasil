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
	"sync"
	"sync/atomic"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

// RateLimitConfig holds rate limiter configuration
type RateLimitConfig struct {
	MaxTokens     uint32        // Bucket capacity
	TokensPerFill uint32        // Tokens added per interval
	FillInterval  time.Duration // Refill interval
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	config *RateLimitConfig

	// Token bucket state
	tokens       uint32
	lastFillTime int64 // Unix nano timestamp
	mu           sync.Mutex

	// Statistics
	allowedCount  uint64
	rejectedCount uint64

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	if config == nil {
		config = &RateLimitConfig{
			MaxTokens:     1000,
			TokensPerFill: 100,
			FillInterval:  time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		config:       config,
		tokens:       config.MaxTokens,
		lastFillTime: time.Now().UnixNano(),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start background token refiller
	rl.wg.Add(1)
	go rl.refillTokens()

	return rl
}

// refillTokens periodically refills tokens
func (rl *RateLimiter) refillTokens() {
	defer rl.wg.Done()

	ticker := time.NewTicker(rl.config.FillInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			current := atomic.LoadUint32(&rl.tokens)
			newTokens := current + rl.config.TokensPerFill
			if newTokens > rl.config.MaxTokens {
				newTokens = rl.config.MaxTokens
			}
			atomic.StoreUint32(&rl.tokens, newTokens)
			atomic.StoreInt64(&rl.lastFillTime, time.Now().UnixNano())
			rl.mu.Unlock()

			logger.DebugField("rate limiter: tokens refilled",
				logger.Uint32("tokens", newTokens),
				logger.Uint32("max", rl.config.MaxTokens))
		}
	}
}

// Allow checks if a request is allowed (non-blocking)
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	current := atomic.LoadUint32(&rl.tokens)
	if current > 0 {
		atomic.StoreUint32(&rl.tokens, current-1)
		atomic.AddUint64(&rl.allowedCount, 1)
		return true
	}

	atomic.AddUint64(&rl.rejectedCount, 1)
	logger.DebugField("rate limiter: request rejected, no tokens available")
	return false
}

// Wait waits for a token to become available (blocking with context)
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		// Try to acquire token
		if rl.Allow() {
			return nil
		}

		// Wait for next refill or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(rl.config.FillInterval / 10):
			// Check again after a short delay
			continue
		}
	}
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.cancel()
	rl.wg.Wait()
	logger.InfoField("rate limiter stopped")
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"current_tokens": atomic.LoadUint32(&rl.tokens),
		"max_tokens":     rl.config.MaxTokens,
		"allowed_count":  atomic.LoadUint64(&rl.allowedCount),
		"rejected_count": atomic.LoadUint64(&rl.rejectedCount),
	}
}
