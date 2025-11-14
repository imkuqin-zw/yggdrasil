package client

import (
	"context"
	"testing"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/stats"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xsync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/code"
)

type mockRemoteClient struct {
	address     string
	serviceName string
	closed      bool
	streams     []*mockClientStream
}

func (m *mockRemoteClient) NewStream(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
	mockStream := &mockClientStream{
		ctx:    ctx,
		desc:   desc,
		method: method,
	}
	m.streams = append(m.streams, mockStream)
	return mockStream, nil
}

func (m *mockRemoteClient) Close() error {
	m.closed = true
	return nil
}

type mockClientStream struct {
	ctx      context.Context
	desc     *stream.StreamDesc
	method   string
	closed   bool
	sentMsg  interface{}
	recvMsg  interface{}
	headers  metadata.MD
	trailers metadata.MD
}

func (m *mockClientStream) Context() context.Context {
	return m.ctx
}

func (m *mockClientStream) SendMsg(msg interface{}) error {
	m.sentMsg = msg
	return nil
}

func (m *mockClientStream) RecvMsg(msg interface{}) error {
	m.recvMsg = msg
	return nil
}

func (m *mockClientStream) Header() (metadata.MD, error) {
	return m.headers, nil
}

func (m *mockClientStream) Trailer() metadata.MD {
	return m.trailers
}

func (m *mockClientStream) CloseSend() error {
	m.closed = true
	return nil
}

type mockResolver struct {
	watches []string
}

func (m *mockResolver) AddWatch(serviceName string) error {
	m.watches = append(m.watches, serviceName)
	return nil
}

func (m *mockResolver) DelWatch(serviceName string) error {
	for i, name := range m.watches {
		if name == serviceName {
			m.watches = append(m.watches[:i], m.watches[i+1:]...)
			break
		}
	}
	return nil
}

type mockBalancer struct {
	name    string
	picker  balancer.Picker
	updated bool
	closed  bool
	values  config.Values
}

func (m *mockBalancer) GetPicker() balancer.Picker {
	return m.picker
}

func (m *mockBalancer) Update(values config.Values) {
	m.values = values
	m.updated = true
}

func (m *mockBalancer) Close() error {
	m.closed = true
	return nil
}

func (m *mockBalancer) Name() string {
	return m.name
}

type mockPicker struct {
	endpoint resolver.Endpoint
	err      error
}

func (m *mockPicker) Next(info balancer.RpcInfo) (balancer.PickResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &mockPickResult{endpoint: m.endpoint}, nil
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

type mockStatsHandler struct {
	stats.Handler
}

func TestInstance(t *testing.T) {
	inst := instance{
		Address:  "localhost:8080",
		Protocol: "grpc",
		Metadata: map[string]interface{}{"weight": 100},
	}

	assert.Equal(t, "localhost:8080", inst.GetAddress())
	assert.Equal(t, "grpc", inst.GetProtocol())
	assert.Equal(t, map[string]interface{}{"weight": 100}, inst.GetMetadata())
}

func TestClientStream(t *testing.T) {
	baseStream := &mockClientStream{
		desc:     &stream.StreamDesc{ServerStreams: false},
		headers:  metadata.MD{"header": []string{"value"}},
		trailers: metadata.MD{"trailer": []string{"value"}},
	}

	cs := &clientStream{
		desc:         &stream.StreamDesc{ServerStreams: false},
		ClientStream: baseStream,
	}

	// Test SendMsg
	err := cs.SendMsg("test message")
	assert.NoError(t, err)
	assert.Equal(t, "test message", baseStream.sentMsg)

	// Test RecvMsg with ServerStreams=false
	var reply interface{}
	err = cs.RecvMsg(&reply)
	assert.NoError(t, err)
	assert.Equal(t, &reply, baseStream.recvMsg)

	// Test RecvMsg with EOF
	baseStream.desc = &stream.StreamDesc{ServerStreams: true}
	err = cs.RecvMsg(&reply)
	assert.NoError(t, err)
}

func TestClient_NewClient_InvalidBalancer(t *testing.T) {
	ctx := context.Background()

	// This will fail because we don't have a balancer registered
	_, err := NewClient(ctx, "test-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found balancer builder")
}

func TestClient_initResolverAndBalancer(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		ctx:           ctx,
		serviceName:   "test-service",
		remoteCli:     make(map[string]remote.Client),
		resolvedEvent: xsync.NewEvent(),
	}

	// Register mock balancer
	balancer.RegisterBuilder("test", func(serviceName string) balancer.Balancer {
		return &mockBalancer{name: serviceName}
	})

	// Test with valid config
	cfg := config.Values{
		"balancer": config.NewValue("test"),
	}

	err := cli.initResolverAndBalancer(cfg)
	require.NoError(t, err)
	assert.NotNil(t, cli.balancer)
	assert.Equal(t, "test-service", cli.balancer.Name())
}

func TestClient_handlePickConfig(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		ctx:           ctx,
		serviceName:   "test-service",
		remoteCli:     make(map[string]remote.Client),
		resolvedEvent: xsync.NewEvent(),
	}

	// Register mock balancer
	mockBal := &mockBalancer{name: "test"}
	cli.balancer = mockBal

	// Register mock remote client builder
	remote.RegisterClientBuilder("grpc", func(ctx context.Context, serviceName string, instance resolver.InsInfo, handler stats.Handler) remote.Client {
		return &mockRemoteClient{address: instance.GetAddress(), serviceName: serviceName}
	})

	// Test with endpoints
	endpoints := []instance{
		{Address: "localhost:8080", Protocol: "grpc"},
		{Address: "localhost:8081", Protocol: "grpc"},
	}

	cfg := config.Values{
		"endpoints": config.NewValue(endpoints),
	}

	cli.handlePickConfig(cfg)

	assert.Equal(t, 2, len(cli.remoteCli))
	assert.True(t, mockBal.updated)
	assert.True(t, cli.resolvedEvent.HasFired())
}

func TestClient_waitForResolved(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		ctx:           ctx,
		resolvedEvent: xsync.NewEvent(),
	}

	// Test already resolved
	cli.resolvedEvent.Fire()
	err := cli.waitForResolved(ctx)
	assert.NoError(t, err)

	// Test timeout
	cli.resolvedEvent = xsync.NewEvent()
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	err = cli.waitForResolved(timeoutCtx)
	assert.Error(t, err)
	assert.Equal(t, code.Code_DEADLINE_EXCEEDED, status.Code(err))
}

func TestClient_newConnStream(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		serviceName: "test-service",
		remoteCli: map[string]remote.Client{
			"localhost:8080": &mockRemoteClient{},
		},
	}

	picker := &mockPicker{
		endpoint: &instance{Address: "localhost:8080"},
	}

	snap := pickSnap{
		balancer:  &mockBalancer{},
		remoteCli: cli.remoteCli,
	}

	desc := &stream.StreamDesc{}

	stream, err := cli.newConnStream(ctx, picker, snap, desc, "TestMethod")
	assert.NoError(t, err)
	assert.NotNil(t, stream)
}

func TestClient_newConnStream_NoClient(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		serviceName: "test-service",
		remoteCli:   make(map[string]remote.Client),
	}

	picker := &mockPicker{
		endpoint: &instance{Address: "localhost:8080"},
	}

	snap := pickSnap{
		balancer:  &mockBalancer{},
		remoteCli: cli.remoteCli,
	}

	desc := &stream.StreamDesc{}

	_, err := cli.newConnStream(ctx, picker, snap, desc, "TestMethod")
	assert.Error(t, err)
	assert.Equal(t, code.Code_UNAVAILABLE, status.Code(err))
}

func TestClient_newStream(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		ctx:           ctx,
		serviceName:   "test-service",
		remoteCli:     make(map[string]remote.Client),
		resolvedEvent: xsync.NewEvent(),
	}

	// Fire resolved event
	cli.resolvedEvent.Fire()

	mockBal := &mockBalancer{
		picker: &mockPicker{
			endpoint: &instance{Address: "localhost:8080"},
		},
	}
	cli.balancer = mockBal
	cli.pickSnap = pickSnap{
		balancer: mockBal,
		remoteCli: map[string]remote.Client{
			"localhost:8080": &mockRemoteClient{},
		},
	}

	desc := &stream.StreamDesc{}

	stream, err := cli.newStream(ctx, desc, "TestMethod")
	assert.NoError(t, err)
	assert.NotNil(t, stream)
}

func TestClient_Invoke(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		ctx:           ctx,
		serviceName:   "test-service",
		remoteCli:     make(map[string]remote.Client),
		resolvedEvent: xsync.NewEvent(),
		unaryInterceptor: func(ctx context.Context, method string, args, reply interface{}, invoker interceptor.Invoker) error {
			return invoker(ctx, method, args, reply)
		},
	}

	// Fire resolved event
	cli.resolvedEvent.Fire()

	mockBal := &mockBalancer{
		picker: &mockPicker{
			endpoint: &instance{Address: "localhost:8080"},
		},
	}
	cli.balancer = mockBal
	cli.pickSnap = pickSnap{
		balancer: mockBal,
		remoteCli: map[string]remote.Client{
			"localhost:8080": &mockRemoteClient{},
		},
	}

	args := "test args"
	reply := ""

	err := cli.Invoke(ctx, "TestMethod", args, &reply)
	assert.NoError(t, err)
}

func TestClient_NewStream(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		ctx:           ctx,
		serviceName:   "test-service",
		remoteCli:     make(map[string]remote.Client),
		resolvedEvent: xsync.NewEvent(),
		streamInterceptor: func(ctx context.Context, desc *stream.StreamDesc, method string, streamer func(context.Context, *stream.StreamDesc, string) (stream.ClientStream, error)) (stream.ClientStream, error) {
			return streamer(ctx, desc, method)
		},
	}

	// Fire resolved event
	cli.resolvedEvent.Fire()

	mockBal := &mockBalancer{
		picker: &mockPicker{
			endpoint: &instance{Address: "localhost:8080"},
		},
	}
	cli.balancer = mockBal
	cli.pickSnap = pickSnap{
		balancer: mockBal,
		remoteCli: map[string]remote.Client{
			"localhost:8080": &mockRemoteClient{},
		},
	}

	desc := &stream.StreamDesc{}

	stream, err := cli.NewStream(ctx, desc, "TestMethod")
	assert.NoError(t, err)
	assert.NotNil(t, stream)
}

func TestClient_Close(t *testing.T) {
	ctx := context.Background()
	cli := &client{
		ctx:           ctx,
		serviceName:   "test-service",
		remoteCli:     make(map[string]remote.Client),
		resolvedEvent: xsync.NewEvent(),
		resolver:      &mockResolver{},
	}

	err := cli.Close()
	assert.NoError(t, err)
}
