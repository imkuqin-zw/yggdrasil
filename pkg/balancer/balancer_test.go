package balancer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEndpoint struct {
	address  string
	protocol string
	metadata map[string]string
}

func (m *mockEndpoint) GetAddress() string {
	return m.address
}

func (m *mockEndpoint) GetProtocol() string {
	return m.protocol
}

func (m *mockEndpoint) GetMetadata() map[string]interface{} {
	md := make(map[string]interface{})
	for k, v := range m.metadata {
		md[k] = v
	}
	return md
}

type mockPickResult struct {
	endpoint resolver.Endpoint
	reported error
}

func (m *mockPickResult) Endpoint() resolver.Endpoint {
	return m.endpoint
}

func (m *mockPickResult) Report(err error) {
	m.reported = err
}

type mockPicker struct {
	endpoints []resolver.Endpoint
	callCount int
}

func (m *mockPicker) Next(info RpcInfo) (PickResult, error) {
	m.callCount++
	if len(m.endpoints) == 0 {
		return nil, ErrNoAvailableInstance
	}

	// Simple round-robin for testing
	endpoint := m.endpoints[m.callCount%len(m.endpoints)]
	return &mockPickResult{endpoint: endpoint}, nil
}

type mockBalancer struct {
	picker    Picker
	name      string
	instances config.Values
	closed    bool
}

func (m *mockBalancer) GetPicker() Picker {
	return m.picker
}

func (m *mockBalancer) Update(values config.Values) {
	m.instances = values
}

func (m *mockBalancer) Close() error {
	m.closed = true
	return nil
}

func (m *mockBalancer) Name() string {
	return m.name
}

func TestGetBuilder(t *testing.T) {
	// Test getting non-existent builder
	_, err := GetBuilder("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found balancer builder")
	assert.Contains(t, err.Error(), "non-existent")

	// Register a builder
	mockBuilder := func(serviceName string) Balancer {
		return &mockBalancer{name: serviceName}
	}
	RegisterBuilder("test", mockBuilder)

	// Test getting registered builder
	builder, err := GetBuilder("test")
	require.NoError(t, err)
	require.NotNil(t, builder)

	// Test creating balancer with builder
	balancer := builder("test-service")
	assert.Equal(t, "test-service", balancer.Name())
}

func TestRegisterBuilder(t *testing.T) {
	// Register a builder
	mockBuilder := func(serviceName string) Balancer {
		return &mockBalancer{name: serviceName}
	}
	RegisterBuilder("register-test", mockBuilder)

	// Verify it was registered
	builder, err := GetBuilder("register-test")
	require.NoError(t, err)
	require.NotNil(t, builder)

	// Test overriding existing builder
	overrideBuilder := func(serviceName string) Balancer {
		return &mockBalancer{name: "override-" + serviceName}
	}
	RegisterBuilder("register-test", overrideBuilder)

	builder, err = GetBuilder("register-test")
	require.NoError(t, err)
	require.NotNil(t, builder)

	balancer := builder("test")
	assert.Equal(t, "override-test", balancer.Name())
}

func TestRpcInfo(t *testing.T) {
	ctx := context.Background()
	method := "TestMethod"

	info := RpcInfo{
		Ctx:    ctx,
		Method: method,
	}

	assert.Equal(t, ctx, info.Ctx)
	assert.Equal(t, method, info.Method)
}

func TestMockEndpoint(t *testing.T) {
	endpoint := &mockEndpoint{
		address:  "localhost:8080",
		protocol: "http",
		metadata: map[string]string{"weight": "100"},
	}

	assert.Equal(t, "localhost:8080", endpoint.GetAddress())
	assert.Equal(t, "http", endpoint.GetProtocol())
	assert.Equal(t, map[string]interface{}{"weight": "100"}, endpoint.GetMetadata())
}

func TestMockPicker(t *testing.T) {
	endpoints := []resolver.Endpoint{
		&mockEndpoint{address: "localhost:8081", protocol: "http"},
		&mockEndpoint{address: "localhost:8082", protocol: "http"},
	}

	picker := &mockPicker{endpoints: endpoints}

	// Test successful picks
	info := RpcInfo{Ctx: context.Background(), Method: "test"}

	result1, err := picker.Next(info)
	require.NoError(t, err)
	require.NotNil(t, result1)
	// First call uses index 1 (0 + 1 % 2 = 1)
	assert.Equal(t, "localhost:8082", result1.Endpoint().GetAddress())

	result2, err := picker.Next(info)
	require.NoError(t, err)
	require.NotNil(t, result2)
	// Second call uses index 0 (1 + 1 % 2 = 0)
	assert.Equal(t, "localhost:8081", result2.Endpoint().GetAddress())

	// Test round-robin behavior
	result3, err := picker.Next(info)
	require.NoError(t, err)
	// Third call uses index 1 (0 + 1 % 2 = 1)
	assert.Equal(t, "localhost:8082", result3.Endpoint().GetAddress())
}

func TestMockPickerNoEndpoints(t *testing.T) {
	picker := &mockPicker{endpoints: []resolver.Endpoint{}}

	info := RpcInfo{Ctx: context.Background(), Method: "test"}
	result, err := picker.Next(info)

	assert.Nil(t, result)
	assert.Equal(t, ErrNoAvailableInstance, err)
}

func TestMockPickResult(t *testing.T) {
	endpoint := &mockEndpoint{address: "localhost:8080", protocol: "http"}
	result := &mockPickResult{endpoint: endpoint}

	// Test Endpoint method
	assert.Equal(t, endpoint, result.Endpoint())

	// Test Report method
	testError := errors.New("test error")
	result.Report(testError)
	assert.Equal(t, testError, result.reported)
}

func TestMockBalancer(t *testing.T) {
	picker := &mockPicker{}
	balancer := &mockBalancer{
		picker: picker,
		name:   "test-balancer",
	}

	// Test GetPicker
	assert.Equal(t, picker, balancer.GetPicker())

	// Test Name
	assert.Equal(t, "test-balancer", balancer.Name())

	// Test Update (using nil values for simplicity)
	balancer.Update(nil)
	assert.Nil(t, balancer.instances)

	// Test Close
	err := balancer.Close()
	assert.NoError(t, err)
	assert.True(t, balancer.closed)
}

func TestConcurrentRegisterGet(t *testing.T) {
	// Test concurrent access to builder registry
	done := make(chan bool, 2)

	// Goroutine 1: Register builders
	go func() {
		for i := 0; i < 100; i++ {
			name := fmt.Sprintf("builder-%d", i)
			RegisterBuilder(name, func(serviceName string) Balancer {
				return &mockBalancer{name: serviceName}
			})
		}
		done <- true
	}()

	// Goroutine 2: Get builders
	go func() {
		for i := 0; i < 100; i++ {
			name := fmt.Sprintf("builder-%d", i)
			if i%10 == 0 { // Try to get some builders that might not exist yet
				GetBuilder(name)
			}
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify at least some builders were registered
	builder, err := GetBuilder("builder-99")
	assert.NoError(t, err)
	assert.NotNil(t, builder)
}

func TestErrorMessages(t *testing.T) {
	// Test ErrNoAvailableInstance
	assert.Equal(t, "no available instance", ErrNoAvailableInstance.Error())

	// Test GetBuilder error message
	_, err := GetBuilder("unknown-balancer")
	expectedError := "not found balancer builder, name: unknown-balancer"
	assert.Equal(t, expectedError, err.Error())
}
