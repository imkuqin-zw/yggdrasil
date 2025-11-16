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

// Tests for the round_robin balancer implementation

func TestInstance(t *testing.T) {
	tests := []struct {
		name     string
		instance *instance
	}{
		{
			name: "basic instance",
			instance: &instance{
				Address:  "localhost:8080",
				Protocol: "http",
				Metadata: map[string]interface{}{"weight": 100, "region": "us-west"},
			},
		},
		{
			name: "instance with nil metadata",
			instance: &instance{
				Address:  "localhost:8081",
				Protocol: "https",
				Metadata: nil,
			},
		},
		{
			name: "instance with empty metadata",
			instance: &instance{
				Address:  "localhost:8082",
				Protocol: "grpc",
				Metadata: map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.instance.Address, tt.instance.GetAddress())
			assert.Equal(t, tt.instance.Protocol, tt.instance.GetProtocol())
			assert.Equal(t, tt.instance.Metadata, tt.instance.GetMetadata())
		})
	}
}

func TestPickResult(t *testing.T) {
	endpoint := &instance{
		Address:  "localhost:8080",
		Protocol: "http",
		Metadata: map[string]interface{}{"test": "value"},
	}

	result := &pickResult{
		ctx:      context.Background(),
		endpoint: endpoint,
	}

	// Test Endpoint method
	assert.Equal(t, endpoint, result.Endpoint())

	// Test Report method (should do nothing)
	result.Report(errors.New("test error"))
	// No assertion needed as Report is empty
}

func TestRoundRobinPicker_Next(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []*instance
		wantErr   bool
	}{
		{
			name:      "no endpoints",
			endpoints: []*instance{},
			wantErr:   true,
		},
		{
			name: "single endpoint",
			endpoints: []*instance{
				{Address: "localhost:8080", Protocol: "http"},
			},
			wantErr: false,
		},
		{
			name: "multiple endpoints",
			endpoints: []*instance{
				{Address: "localhost:8081", Protocol: "http"},
				{Address: "localhost:8082", Protocol: "http"},
				{Address: "localhost:8083", Protocol: "http"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker := &roundRobinPicker{
				endpoint: tt.endpoints,
			}

			info := RpcInfo{
				Ctx:    context.Background(),
				Method: "TestMethod",
			}

			if tt.wantErr {
				result, err := picker.Next(info)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "not found endpoint")
			} else {
				// Test multiple picks to verify round-robin behavior
				seenAddresses := make(map[string]bool)
				for i := 0; i < len(tt.endpoints)*2; i++ {
					result, err := picker.Next(info)
					assert.NoError(t, err)
					assert.NotNil(t, result)

					endpoint := result.Endpoint()
					assert.NotNil(t, endpoint)

					address := endpoint.GetAddress()
					seenAddresses[address] = true
				}

				// Should have seen all endpoints at least once
				if len(tt.endpoints) > 0 {
					for _, endpoint := range tt.endpoints {
						assert.True(t, seenAddresses[endpoint.Address],
							"should have seen endpoint %s", endpoint.Address)
					}
				}
			}
		})
	}
}

func TestNewRoundRobin(t *testing.T) {
	balancer := newRoundRobin("test-service")
	assert.NotNil(t, balancer)
	assert.Equal(t, "round_robin", balancer.Name())
	rrBalancer := balancer.(*RoundRobin)
	assert.Equal(t, int64(0), rrBalancer.idx.Load())
	assert.Empty(t, rrBalancer.endpoint)
}

func TestRoundRobin_GetPicker(t *testing.T) {
	balancer := &RoundRobin{
		endpoint: []*instance{
			{Address: "localhost:8080", Protocol: "http"},
			{Address: "localhost:8081", Protocol: "http"},
		},
	}

	// Test getting multiple pickers
	picker1 := balancer.GetPicker()
	assert.NotNil(t, picker1)

	picker2 := balancer.GetPicker()
	assert.NotNil(t, picker2)
	assert.NotEqual(t, picker1, picker2) // Should be different instances

	// Verify pickers have the same endpoints
	rrPicker1 := picker1.(*roundRobinPicker)
	rrPicker2 := picker2.(*roundRobinPicker)
	assert.Equal(t, len(balancer.endpoint), len(rrPicker1.endpoint))
	assert.Equal(t, len(balancer.endpoint), len(rrPicker2.endpoint))
}

func TestRoundRobin_Update(t *testing.T) {
	balancer := &RoundRobin{}

	// Test update with endpoints directly
	testEndpoints := []*instance{
		{Address: "localhost:8080", Protocol: "http"},
		{Address: "localhost:8081", Protocol: "https"},
	}

	// Create a simple implementation for testing
	balancer.updateWithEndpoints(testEndpoints)
	assert.Len(t, balancer.endpoint, 2)
	assert.Equal(t, "localhost:8080", balancer.endpoint[0].Address)
	assert.Equal(t, "localhost:8081", balancer.endpoint[1].Address)

	// Test update with empty endpoints
	balancer.updateWithEndpoints([]*instance{})
	assert.Empty(t, balancer.endpoint)
}

// Helper method to update endpoints directly for testing
func (b *RoundRobin) updateWithEndpoints(endpoints []*instance) {
	b.endpoint = endpoints
}

func TestRoundRobin_Close(t *testing.T) {
	balancer := &RoundRobin{
		endpoint: []*instance{
			{Address: "localhost:8080", Protocol: "http"},
		},
	}

	err := balancer.Close()
	assert.NoError(t, err)
	// Balancer should still be functional after close (no cleanup required)
	assert.NotEmpty(t, balancer.endpoint)
}

func TestRoundRobin_Name(t *testing.T) {
	balancer := &RoundRobin{}
	assert.Equal(t, "round_robin", balancer.Name())
}

func TestRoundRobinIntegration(t *testing.T) {
	// Test the complete round-robin flow
	balancer := newRoundRobin("test-service")

	// Setup test endpoints
	endpoints := []*instance{
		{Address: "localhost:8081", Protocol: "http", Metadata: map[string]interface{}{"weight": 1}},
		{Address: "localhost:8082", Protocol: "http", Metadata: map[string]interface{}{"weight": 2}},
		{Address: "localhost:8083", Protocol: "https", Metadata: map[string]interface{}{"weight": 3}},
	}
	rrBalancer := balancer.(*RoundRobin)
	rrBalancer.updateWithEndpoints(endpoints)

	// Get picker and test multiple selections
	picker := balancer.GetPicker()

	info := RpcInfo{
		Ctx:    context.Background(),
		Method: "TestService.TestMethod",
	}

	// Test round-robin selection
	selectedAddresses := make([]string, 6)
	for i := 0; i < 6; i++ {
		result, err := picker.Next(info)
		require.NoError(t, err)
		require.NotNil(t, result)

		endpoint := result.Endpoint()
		selectedAddresses[i] = endpoint.GetAddress()
	}

	// Verify round-robin behavior: should cycle through endpoints
	// Note: picker starts with idx=1 due to b.idx.Add(1) in GetPicker()
	expectedOrder := []string{
		"localhost:8082", "localhost:8083", "localhost:8081",
		"localhost:8082", "localhost:8083", "localhost:8081",
	}
	assert.Equal(t, expectedOrder, selectedAddresses)
}

func TestRoundRobinWithConcurrentAccess(t *testing.T) {
	balancer := &RoundRobin{}
	endpoints := make([]*instance, 10)
	for i := 0; i < 10; i++ {
		endpoints[i] = &instance{
			Address:  fmt.Sprintf("localhost:%d", 8080+i),
			Protocol: "http",
		}
	}
	balancer.updateWithEndpoints(endpoints)

	// Test concurrent picker creation and usage
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			picker := balancer.GetPicker()
			info := RpcInfo{Ctx: context.Background(), Method: "test"}

			// Each goroutine makes several picks
			for j := 0; j < 5; j++ {
				result, err := picker.Next(info)
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Endpoint())
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
