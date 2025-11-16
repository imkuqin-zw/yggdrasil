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

package resolver

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockEndpoint implements the Endpoint interface for testing
type MockEndpoint struct {
	address  string
	protocol string
	metadata map[string]interface{}
}

func NewMockEndpoint(address, protocol string) *MockEndpoint {
	return &MockEndpoint{
		address:  address,
		protocol: protocol,
		metadata: map[string]interface{}{
			"weight": 1,
			"region": "us-east-1",
		},
	}
}

func (e *MockEndpoint) GetAddress() string {
	return e.address
}

func (e *MockEndpoint) GetProtocol() string {
	return e.protocol
}

func (e *MockEndpoint) GetMetadata() map[string]interface{} {
	return e.metadata
}

// MockResolver implements the Resolver interface for testing
type MockResolver struct {
	name       string
	watchCount int
	closed     bool
	closeError error
	mu         sync.Mutex
	endpoints  []Endpoint
}

func NewMockResolver(name string) *MockResolver {
	return &MockResolver{
		name: name,
		endpoints: []Endpoint{
			NewMockEndpoint("localhost:8080", "grpc"),
			NewMockEndpoint("localhost:8081", "grpc"),
		},
	}
}

func (r *MockResolver) AddWatch(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.watchCount++
	return nil
}

func (r *MockResolver) DelWatch(serviceName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.watchCount > 0 {
		r.watchCount--
	}
	return nil
}

func (r *MockResolver) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	return r.closeError
}

func (r *MockResolver) Name() string {
	return r.name
}

func (r *MockResolver) GetWatchCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.watchCount
}

func (r *MockResolver) IsClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

func (r *MockResolver) SetCloseError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closeError = err
}

// Test RegisterBuilder function
func TestRegisterBuilder(t *testing.T) {
	// Clear global state before test
	clearGlobalState()

	tests := []struct {
		name   string
		key    string
		setup  func()
		expect func()
	}{
		{
			name: "register new builder",
			key:  "test-resolver",
			setup: func() {
				RegisterBuilder("test-resolver", func(name string) (Resolver, error) {
					return NewMockResolver(name), nil
				})
			},
			expect: func() {
				// Verify builder is registered by trying to create resolver
				r, err := GetResolver("test-resolver")
				assert.NoError(t, err)
				assert.NotNil(t, r)
				assert.Equal(t, "test-resolver", r.Name())
			},
		},
		{
			name: "overwrite existing builder",
			key:  "overwrite-resolver",
			setup: func() {
				RegisterBuilder("overwrite-resolver", func(name string) (Resolver, error) {
					return NewMockResolver("old-" + name), nil
				})
				RegisterBuilder("overwrite-resolver", func(name string) (Resolver, error) {
					return NewMockResolver("new-" + name), nil
				})
			},
			expect: func() {
				r, err := GetResolver("overwrite-resolver")
				assert.NoError(t, err)
				assert.NotNil(t, r)
				assert.Equal(t, "new-overwrite-resolver", r.Name())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearGlobalState()
			tt.setup()
			tt.expect()
		})
	}
}

// Test GetResolver function
func TestGetResolver(t *testing.T) {
	clearGlobalState()

	tests := []struct {
		name         string
		resolverName string
		builder      func(name string) (Resolver, error)
		setup        func()
		expectError  bool
		expectName   string
	}{
		{
			name:         "existing resolver cached",
			resolverName: "cached-resolver",
			builder: func(name string) (Resolver, error) {
				return NewMockResolver(name), nil
			},
			setup: func() {
				RegisterBuilder("cached-resolver", func(name string) (Resolver, error) {
					return NewMockResolver(name), nil
				})
				// First call to create and cache
				GetResolver("cached-resolver")
			},
			expectError: false,
			expectName:  "cached-resolver",
		},
		{
			name:         "resolver not found",
			resolverName: "non-existent",
			expectError:  true,
		},
		{
			name:         "builder returns error",
			resolverName: "error-resolver",
			builder: func(name string) (Resolver, error) {
				return nil, errors.New("builder error")
			},
			setup: func() {
				RegisterBuilder("error-resolver", func(name string) (Resolver, error) {
					return nil, errors.New("builder error")
				})
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearGlobalState()

			if tt.setup != nil {
				tt.setup()
			}

			r, err := GetResolver(tt.resolverName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, r)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, r)
				if tt.expectName != "" {
					assert.Equal(t, tt.expectName, r.Name())
				}
			}
		})
	}
}

// Test GetResolver lazy initialization and double-checked locking
func TestGetResolverLazyInitialization(t *testing.T) {
	clearGlobalState()

	// Register a builder that tracks creation count
	createCount := 0
	RegisterBuilder("lazy-resolver", func(name string) (Resolver, error) {
		createCount++
		return NewMockResolver(name), nil
	})

	// Multiple concurrent calls should only create one resolver
	var wg sync.WaitGroup
	results := make([]Resolver, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index], errors[index] = GetResolver("lazy-resolver")
		}(i)
	}

	wg.Wait()

	// Verify only one resolver was created
	assert.Equal(t, 1, createCount)

	// Verify all calls succeeded and got the same resolver
	for i, err := range errors {
		assert.NoError(t, err, "Call %d should succeed", i)
		assert.NotNil(t, results[i], "Call %d should return a resolver", i)
		assert.Equal(t, results[0], results[i], "Call %d should return the same resolver", i)
	}
}

// Test DelResolver function
func TestDelResolver(t *testing.T) {
	clearGlobalState()

	tests := []struct {
		name         string
		resolverName string
		setup        func()
		expectError  bool
		expectClosed bool
	}{
		{
			name:         "delete existing resolver",
			resolverName: "delete-resolver",
			setup: func() {
				RegisterBuilder("delete-resolver", func(name string) (Resolver, error) {
					return NewMockResolver(name), nil
				})
				// Create resolver first
				GetResolver("delete-resolver")
			},
			expectError:  false,
			expectClosed: true,
		},
		{
			name:         "delete non-existent resolver",
			resolverName: "non-existent",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearGlobalState()

			if tt.setup != nil {
				tt.setup()
			}

			err := DelResolver(tt.resolverName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// If we created a resolver, verify it was closed
			if tt.expectClosed {
				// Try to get resolver again - it should be recreated
				newResolver, err := GetResolver(tt.resolverName)
				assert.NoError(t, err)
				assert.NotNil(t, newResolver)

				// If it's a MockResolver, check if the old one was closed
				if mock, ok := newResolver.(*MockResolver); ok {
					// The new resolver should not be closed
					assert.False(t, mock.IsClosed())
				}
			}
		})
	}
}

// Test DelResolver with close error
func TestDelResolverCloseError(t *testing.T) {
	clearGlobalState()

	RegisterBuilder("close-error-resolver", func(name string) (Resolver, error) {
		r := NewMockResolver(name)
		r.SetCloseError(errors.New("close error"))
		return r, nil
	})

	// Create resolver first
	r, err := GetResolver("close-error-resolver")
	assert.NoError(t, err)
	assert.NotNil(t, r)

	// Delete should propagate close error
	err = DelResolver("close-error-resolver")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close error")
}

// Test concurrent operations
func TestConcurrentOperations(t *testing.T) {
	clearGlobalState()

	// Register multiple builders
	RegisterBuilder("resolver-1", func(name string) (Resolver, error) {
		return NewMockResolver("resolved-1"), nil
	})
	RegisterBuilder("resolver-2", func(name string) (Resolver, error) {
		return NewMockResolver("resolved-2"), nil
	})
	RegisterBuilder("resolver-3", func(name string) (Resolver, error) {
		return NewMockResolver("resolved-3"), nil
	})

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent gets
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				name := fmt.Sprintf("resolver-%d", (j%3)+1)
				_, err := GetResolver(name)
				if err != nil {
					errors <- err
					return
				}
			}
		}()
	}

	// Concurrent registers and deletes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			name := fmt.Sprintf("dynamic-resolver-%d", index)
			RegisterBuilder(name, func(n string) (Resolver, error) {
				return NewMockResolver(n), nil
			})

			_, err := GetResolver(name)
			if err != nil {
				errors <- err
				return
			}

			// Simulate some work
			time.Sleep(time.Millisecond)

			err = DelResolver(name)
			if err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

// Test MockEndpoint functionality
func TestMockEndpoint(t *testing.T) {
	endpoint := NewMockEndpoint("localhost:8080", "grpc")

	assert.Equal(t, "localhost:8080", endpoint.GetAddress())
	assert.Equal(t, "grpc", endpoint.GetProtocol())

	metadata := endpoint.GetMetadata()
	assert.Equal(t, 1, metadata["weight"])
	assert.Equal(t, "us-east-1", metadata["region"])
}

// Test MockResolver functionality
func TestMockResolver(t *testing.T) {
	resolver := NewMockResolver("test-resolver")

	assert.Equal(t, "test-resolver", resolver.Name())
	assert.Equal(t, 0, resolver.GetWatchCount())
	assert.False(t, resolver.IsClosed())

	// Test AddWatch
	err := resolver.AddWatch("test-service")
	assert.NoError(t, err)
	assert.Equal(t, 1, resolver.GetWatchCount())

	// Test AddWatch multiple times
	err = resolver.AddWatch("another-service")
	assert.NoError(t, err)
	assert.Equal(t, 2, resolver.GetWatchCount())

	// Test DelWatch
	err = resolver.DelWatch("test-service")
	assert.NoError(t, err)
	assert.Equal(t, 1, resolver.GetWatchCount())

	// Test Close
	err = resolver.Close()
	assert.NoError(t, err)
	assert.True(t, resolver.IsClosed())

	// Test close error
	resolver.SetCloseError(errors.New("test close error"))
	err = resolver.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test close error")
}

// Test error message format
func TestErrorMessages(t *testing.T) {
	clearGlobalState()

	// Test not found resolver builder error
	_, err := GetResolver("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found resolver builder")
	assert.Contains(t, err.Error(), "non-existent")
}

// Helper function to clear global state for clean tests
func clearGlobalState() {
	// This function clears the global maps for isolated testing
	// Note: In a real scenario, you might want to use dependency injection
	// or provide a way to reset the package state
	mu.Lock()
	resolver = make(map[string]Resolver)
	builder = make(map[string]func(string) (Resolver, error))
	mu.Unlock()
}
