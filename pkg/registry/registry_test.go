package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRegistry struct {
	name string
}

func (m *mockRegistry) Register(ctx context.Context, instance Instance) error {
	return nil
}

func (m *mockRegistry) Deregister(ctx context.Context, instance Instance) error {
	return nil
}

func (m *mockRegistry) Name() string {
	return m.name
}

type mockEndpoint struct {
	scheme   string
	address  string
	metadata map[string]string
}

func (m *mockEndpoint) Scheme() string {
	return m.scheme
}

func (m *mockEndpoint) Address() string {
	return m.address
}

func (m *mockEndpoint) Metadata() map[string]string {
	return m.metadata
}

type mockInstance struct {
	region    string
	zone      string
	campus    string
	namespace string
	name      string
	version   string
	metadata  map[string]string
	endpoints []Endpoint
}

func (m *mockInstance) Region() string {
	return m.region
}

func (m *mockInstance) Zone() string {
	return m.zone
}

func (m *mockInstance) Campus() string {
	return m.campus
}

func (m *mockInstance) Namespace() string {
	return m.namespace
}

func (m *mockInstance) Name() string {
	return m.name
}

func (m *mockInstance) Version() string {
	return m.version
}

func (m *mockInstance) Metadata() map[string]string {
	return m.metadata
}

func (m *mockInstance) Endpoints() []Endpoint {
	return m.endpoints
}

func TestRegisterBuilder(t *testing.T) {
	name := "test-registry"
	builder := func() Registry {
		return &mockRegistry{name: name}
	}

	// Register builder
	RegisterBuilder(name, builder)

	// Verify builder was registered
	retrievedBuilder := GetBuilder(name)
	require.NotNil(t, retrievedBuilder)

	// Test creating registry with builder
	registry := retrievedBuilder()
	assert.NotNil(t, registry)
	assert.Equal(t, name, registry.Name())
}

func TestGetBuilder_NonExistent(t *testing.T) {
	// Test getting non-existent builder
	builder := GetBuilder("non-existent-registry")
	assert.Nil(t, builder)
}

func TestRegisterBuilder_Override(t *testing.T) {
	name := "override-registry"

	// Register initial builder
	initialBuilder := func() Registry {
		return &mockRegistry{name: "initial"}
	}
	RegisterBuilder(name, initialBuilder)

	// Override with new builder
	overrideBuilder := func() Registry {
		return &mockRegistry{name: "override"}
	}
	RegisterBuilder(name, overrideBuilder)

	// Verify we get the override builder
	retrievedBuilder := GetBuilder(name)
	require.NotNil(t, retrievedBuilder)

	registry := retrievedBuilder()
	assert.Equal(t, "override", registry.Name())
}

func TestMockRegistry(t *testing.T) {
	registry := &mockRegistry{name: "test-registry"}

	ctx := context.Background()
	instance := &mockInstance{}

	// Test Register
	err := registry.Register(ctx, instance)
	assert.NoError(t, err)

	// Test Deregister
	err = registry.Deregister(ctx, instance)
	assert.NoError(t, err)

	// Test Name
	assert.Equal(t, "test-registry", registry.Name())
}

func TestMockEndpoint(t *testing.T) {
	endpoint := &mockEndpoint{
		scheme:   "http",
		address:  "localhost:8080",
		metadata: map[string]string{"weight": "100"},
	}

	assert.Equal(t, "http", endpoint.Scheme())
	assert.Equal(t, "localhost:8080", endpoint.Address())
	assert.Equal(t, map[string]string{"weight": "100"}, endpoint.Metadata())
}

func TestMockInstance(t *testing.T) {
	endpoints := []Endpoint{
		&mockEndpoint{
			scheme:   "http",
			address:  "localhost:8080",
			metadata: map[string]string{"weight": "100"},
		},
		&mockEndpoint{
			scheme:   "grpc",
			address:  "localhost:9090",
			metadata: map[string]string{"weight": "200"},
		},
	}

	instance := &mockInstance{
		region:    "us-west-1",
		zone:      "us-west-1a",
		campus:    "campus1",
		namespace: "default",
		name:      "test-service",
		version:   "v1.0.0",
		metadata:  map[string]string{"env": "production"},
		endpoints: endpoints,
	}

	// Test all getter methods
	assert.Equal(t, "us-west-1", instance.Region())
	assert.Equal(t, "us-west-1a", instance.Zone())
	assert.Equal(t, "campus1", instance.Campus())
	assert.Equal(t, "default", instance.Namespace())
	assert.Equal(t, "test-service", instance.Name())
	assert.Equal(t, "v1.0.0", instance.Version())
	assert.Equal(t, map[string]string{"env": "production"}, instance.Metadata())

	// Test Endpoints
	retrievedEndpoints := instance.Endpoints()
	require.Equal(t, 2, len(retrievedEndpoints))

	// Test first endpoint
	assert.Equal(t, "http", retrievedEndpoints[0].Scheme())
	assert.Equal(t, "localhost:8080", retrievedEndpoints[0].Address())
	assert.Equal(t, map[string]string{"weight": "100"}, retrievedEndpoints[0].Metadata())

	// Test second endpoint
	assert.Equal(t, "grpc", retrievedEndpoints[1].Scheme())
	assert.Equal(t, "localhost:9090", retrievedEndpoints[1].Address())
	assert.Equal(t, map[string]string{"weight": "200"}, retrievedEndpoints[1].Metadata())
}

func TestMDServerKind(t *testing.T) {
	assert.Equal(t, "serverKind", MDServerKind)
}

func TestRegistryInterface(t *testing.T) {
	// This test ensures our mock implements the Registry interface
	var _ Registry = &mockRegistry{}
}

func TestEndpointInterface(t *testing.T) {
	// This test ensures our mock implements the Endpoint interface
	var _ Endpoint = &mockEndpoint{}
}

func TestInstanceInterface(t *testing.T) {
	// This test ensures our mock implements the Instance interface
	var _ Instance = &mockInstance{}
}

func TestMultipleRegistryBuilders(t *testing.T) {
	// Register multiple builders
	builders := map[string]string{
		"consul":  "consul-registry",
		"etcd":    "etcd-registry",
		"nacos":   "nacos-registry",
		"polaris": "polaris-registry",
		"eureka":  "eureka-registry",
	}

	for name, registryName := range builders {
		name := name
		registryName := registryName
		builder := func() Registry {
			return &mockRegistry{name: registryName}
		}
		RegisterBuilder(name, builder)
	}

	// Verify all builders were registered
	for name, expectedName := range builders {
		retrievedBuilder := GetBuilder(name)
		require.NotNil(t, retrievedBuilder, "Builder %s should be registered", name)

		registry := retrievedBuilder()
		assert.Equal(t, expectedName, registry.Name(), "Registry name should match expected")
	}
}

func TestBuilderConcurrency(t *testing.T) {
	// Test concurrent registration and retrieval
	done := make(chan bool, 2)

	// Goroutine 1: Register builders
	go func() {
		for i := 0; i < 100; i++ {
			name := "concurrent-registry-" + string(rune(i))
			builder := func() Registry {
				return &mockRegistry{name: name}
			}
			RegisterBuilder(name, builder)
		}
		done <- true
	}()

	// Goroutine 2: Get builders
	go func() {
		for i := 0; i < 100; i++ {
			name := "concurrent-registry-" + string(rune(i))
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
	for i := 0; i < 10; i++ {
		name := "concurrent-registry-" + string(rune(i))
		builder := GetBuilder(name)
		if builder != nil {
			registry := builder()
			assert.Equal(t, name, registry.Name())
			break
		}
	}
}

func TestRegistryIntegration(t *testing.T) {
	// Test full integration with Instance and Endpoint
	registry := &mockRegistry{name: "integration-test"}

	endpoint := &mockEndpoint{
		scheme:   "grpc",
		address:  "localhost:9090",
		metadata: map[string]string{MDServerKind: "grpc"},
	}

	instance := &mockInstance{
		region:    "us-east-1",
		zone:      "us-east-1a",
		campus:    "dc1",
		namespace: "production",
		name:      "integration-service",
		version:   "v2.0.0",
		metadata:  map[string]string{"team": "platform"},
		endpoints: []Endpoint{endpoint},
	}

	ctx := context.Background()

	// Test Register
	err := registry.Register(ctx, instance)
	assert.NoError(t, err)

	// Test Deregister
	err = registry.Deregister(ctx, instance)
	assert.NoError(t, err)

	// Verify the instance and endpoint data
	assert.Equal(t, "us-east-1", instance.Region())
	assert.Equal(t, "grpc", endpoint.Scheme())
	assert.Equal(t, "grpc", endpoint.Metadata()[MDServerKind])
}
