package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClientStream struct {
	ctx context.Context
}

func (m *mockClientStream) Context() context.Context {
	return m.ctx
}

func (m *mockClientStream) SendMsg(interface{}) error {
	return nil
}

func (m *mockClientStream) RecvMsg(interface{}) error {
	return nil
}

func (m *mockClientStream) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *mockClientStream) Trailer() metadata.MD {
	return nil
}

func (m *mockClientStream) CloseSend() error {
	return nil
}

type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) SendMsg(interface{}) error {
	return nil
}

func (m *mockServerStream) RecvMsg(interface{}) error {
	return nil
}

func (m *mockServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetTrailer(metadata.MD) {
}

func TestRegisterUnaryClientIntBuilder(t *testing.T) {
	name := "test-unary-client"
	builder := func(serviceName string) UnaryClientInterceptor {
		return func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error {
			return invoker(ctx, method, req, reply)
		}
	}

	RegisterUnaryClientIntBuilder(name, builder)

	// Test that the builder was registered
	retrievedBuilder := getUnaryClientIntBuilder(name)
	require.NotNil(t, retrievedBuilder)

	// Test that the interceptor works
	interceptor := retrievedBuilder("test-service")
	require.NotNil(t, interceptor)

	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}) error {
		called = true
		return nil
	}

	err := interceptor(context.Background(), "TestMethod", "req", "reply", invoker)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRegisterUnaryServerIntBuilder(t *testing.T) {
	name := "test-unary-server"
	builder := func() UnaryServerInterceptor {
		return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error) {
			return handler(ctx, req)
		}
	}

	RegisterUnaryServerIntBuilder(name, builder)

	// Test that the builder was registered
	retrievedBuilder := getUnaryServerIntBuilder(name)
	require.NotNil(t, retrievedBuilder)

	// Test that the interceptor works
	interceptor := retrievedBuilder()
	require.NotNil(t, interceptor)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	}

	info := &UnaryServerInfo{FullMethod: "/test.Service/Method"}
	resp, err := interceptor(context.Background(), "req", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, called)
}

func TestRegisterStreamClientIntBuilder(t *testing.T) {
	name := "test-stream-client"
	builder := func(serviceName string) StreamClientInterceptor {
		return func(ctx context.Context, desc *stream.StreamDesc, method string, streamer Streamer) (stream.ClientStream, error) {
			return streamer(ctx, desc, method)
		}
	}

	RegisterStreamClientIntBuilder(name, builder)

	// Test that the builder was registered
	retrievedBuilder := getStreamClientIntBuilder(name)
	require.NotNil(t, retrievedBuilder)

	// Test that the interceptor works
	interceptor := retrievedBuilder("test-service")
	require.NotNil(t, interceptor)

	called := false
	streamer := func(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
		called = true
		return &mockClientStream{ctx: ctx}, nil
	}

	desc := &stream.StreamDesc{}
	cs, err := interceptor(context.Background(), desc, "TestMethod", streamer)
	assert.NoError(t, err)
	assert.NotNil(t, cs)
	assert.True(t, called)
}

func TestRegisterStreamServerIntBuilder(t *testing.T) {
	name := "test-stream-server"
	builder := func() StreamServerInterceptor {
		return func(srv interface{}, ss stream.ServerStream, info *StreamServerInfo, handler stream.StreamHandler) error {
			return handler(srv, ss)
		}
	}

	RegisterStreamServerIntBuilder(name, builder)

	// Test that the builder was registered
	retrievedBuilder := getStreamServerIntBuilder(name)
	require.NotNil(t, retrievedBuilder)

	// Test that the interceptor works
	interceptor := retrievedBuilder()
	require.NotNil(t, interceptor)

	called := false
	handler := func(srv interface{}, ss stream.ServerStream) error {
		called = true
		return nil
	}

	info := &StreamServerInfo{FullMethod: "/test.Service/Method"}
	ss := &mockServerStream{ctx: context.Background()}

	err := interceptor("test-service", ss, info, handler)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestChainUnaryClientInterceptors_Empty(t *testing.T) {
	interceptor := ChainUnaryClientInterceptors("test-service", []string{})
	require.NotNil(t, interceptor)

	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}) error {
		called = true
		return nil
	}

	err := interceptor(context.Background(), "TestMethod", "req", "reply", invoker)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestChainUnaryClientInterceptors_Single(t *testing.T) {
	name := "test-single-unary-client"
	RegisterUnaryClientIntBuilder(name, func(serviceName string) UnaryClientInterceptor {
		return func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error {
			return invoker(ctx, method, req, reply)
		}
	})

	interceptor := ChainUnaryClientInterceptors("test-service", []string{name})
	require.NotNil(t, interceptor)

	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}) error {
		called = true
		return nil
	}

	err := interceptor(context.Background(), "TestMethod", "req", "reply", invoker)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestChainUnaryClientInterceptors_Multiple(t *testing.T) {
	name1 := "test-multi-unary-client-1"
	name2 := "test-multi-unary-client-2"

	callOrder := []string{}

	RegisterUnaryClientIntBuilder(name1, func(serviceName string) UnaryClientInterceptor {
		return func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error {
			callOrder = append(callOrder, name1)
			return invoker(ctx, method, req, reply)
		}
	})

	RegisterUnaryClientIntBuilder(name2, func(serviceName string) UnaryClientInterceptor {
		return func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error {
			callOrder = append(callOrder, name2)
			return invoker(ctx, method, req, reply)
		}
	})

	interceptor := ChainUnaryClientInterceptors("test-service", []string{name1, name2})
	require.NotNil(t, interceptor)

	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}) error {
		called = true
		return nil
	}

	err := interceptor(context.Background(), "TestMethod", "req", "reply", invoker)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, []string{name1, name2}, callOrder)
}

func TestChainUnaryClientInterceptors_NonExistent(t *testing.T) {
	// This should not panic, just warn and return a pass-through interceptor
	interceptor := ChainUnaryClientInterceptors("test-service", []string{"non-existent-interceptor"})
	require.NotNil(t, interceptor)

	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}) error {
		called = true
		return nil
	}

	err := interceptor(context.Background(), "TestMethod", "req", "reply", invoker)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestChainStreamClientInterceptors_Empty(t *testing.T) {
	interceptor := ChainStreamClientInterceptors("test-service", []string{})
	require.NotNil(t, interceptor)

	called := false
	streamer := func(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
		called = true
		return &mockClientStream{ctx: ctx}, nil
	}

	desc := &stream.StreamDesc{}
	cs, err := interceptor(context.Background(), desc, "TestMethod", streamer)
	assert.NoError(t, err)
	assert.NotNil(t, cs)
	assert.True(t, called)
}

func TestChainStreamClientInterceptors_Single(t *testing.T) {
	name := "test-single-stream-client"
	RegisterStreamClientIntBuilder(name, func(serviceName string) StreamClientInterceptor {
		return func(ctx context.Context, desc *stream.StreamDesc, method string, streamer Streamer) (stream.ClientStream, error) {
			return streamer(ctx, desc, method)
		}
	})

	interceptor := ChainStreamClientInterceptors("test-service", []string{name})
	require.NotNil(t, interceptor)

	called := false
	streamer := func(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
		called = true
		return &mockClientStream{ctx: ctx}, nil
	}

	desc := &stream.StreamDesc{}
	cs, err := interceptor(context.Background(), desc, "TestMethod", streamer)
	assert.NoError(t, err)
	assert.NotNil(t, cs)
	assert.True(t, called)
}

func TestChainStreamClientInterceptors_Multiple(t *testing.T) {
	name1 := "test-multi-stream-client-1"
	name2 := "test-multi-stream-client-2"

	callOrder := []string{}

	RegisterStreamClientIntBuilder(name1, func(serviceName string) StreamClientInterceptor {
		return func(ctx context.Context, desc *stream.StreamDesc, method string, streamer Streamer) (stream.ClientStream, error) {
			callOrder = append(callOrder, name1)
			return streamer(ctx, desc, method)
		}
	})

	RegisterStreamClientIntBuilder(name2, func(serviceName string) StreamClientInterceptor {
		return func(ctx context.Context, desc *stream.StreamDesc, method string, streamer Streamer) (stream.ClientStream, error) {
			callOrder = append(callOrder, name2)
			return streamer(ctx, desc, method)
		}
	})

	interceptor := ChainStreamClientInterceptors("test-service", []string{name1, name2})
	require.NotNil(t, interceptor)

	called := false
	streamer := func(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
		called = true
		return &mockClientStream{ctx: ctx}, nil
	}

	desc := &stream.StreamDesc{}
	cs, err := interceptor(context.Background(), desc, "TestMethod", streamer)
	assert.NoError(t, err)
	assert.NotNil(t, cs)
	assert.True(t, called)
	assert.Equal(t, []string{name1, name2}, callOrder)
}

func TestChainUnaryServerInterceptors_Empty(t *testing.T) {
	interceptor := ChainUnaryServerInterceptors([]string{})
	require.NotNil(t, interceptor)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	}

	info := &UnaryServerInfo{FullMethod: "/test.Service/Method"}
	resp, err := interceptor(context.Background(), "req", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, called)
}

func TestChainUnaryServerInterceptors_Single(t *testing.T) {
	name := "test-single-unary-server"
	RegisterUnaryServerIntBuilder(name, func() UnaryServerInterceptor {
		return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error) {
			return handler(ctx, req)
		}
	})

	interceptor := ChainUnaryServerInterceptors([]string{name})
	require.NotNil(t, interceptor)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	}

	info := &UnaryServerInfo{FullMethod: "/test.Service/Method"}
	resp, err := interceptor(context.Background(), "req", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, called)
}

func TestChainUnaryServerInterceptors_Multiple(t *testing.T) {
	name1 := "test-multi-unary-server-1"
	name2 := "test-multi-unary-server-2"

	callOrder := []string{}

	RegisterUnaryServerIntBuilder(name1, func() UnaryServerInterceptor {
		return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error) {
			callOrder = append(callOrder, name1)
			return handler(ctx, req)
		}
	})

	RegisterUnaryServerIntBuilder(name2, func() UnaryServerInterceptor {
		return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error) {
			callOrder = append(callOrder, name2)
			return handler(ctx, req)
		}
	})

	interceptor := ChainUnaryServerInterceptors([]string{name1, name2})
	require.NotNil(t, interceptor)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	}

	info := &UnaryServerInfo{FullMethod: "/test.Service/Method"}
	resp, err := interceptor(context.Background(), "req", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, called)
	assert.Equal(t, []string{name1, name2}, callOrder)
}

func TestChainStreamServerInterceptors_Empty(t *testing.T) {
	interceptor := ChainStreamServerInterceptors([]string{})
	require.NotNil(t, interceptor)

	called := false
	handler := func(srv interface{}, ss stream.ServerStream) error {
		called = true
		return nil
	}

	info := &StreamServerInfo{FullMethod: "/test.Service/Method"}
	ss := &mockServerStream{ctx: context.Background()}

	err := interceptor("test-service", ss, info, handler)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestChainStreamServerInterceptors_Single(t *testing.T) {
	name := "test-single-stream-server"
	RegisterStreamServerIntBuilder(name, func() StreamServerInterceptor {
		return func(srv interface{}, ss stream.ServerStream, info *StreamServerInfo, handler stream.StreamHandler) error {
			return handler(srv, ss)
		}
	})

	interceptor := ChainStreamServerInterceptors([]string{name})
	require.NotNil(t, interceptor)

	called := false
	handler := func(srv interface{}, ss stream.ServerStream) error {
		called = true
		return nil
	}

	info := &StreamServerInfo{FullMethod: "/test.Service/Method"}
	ss := &mockServerStream{ctx: context.Background()}

	err := interceptor("test-service", ss, info, handler)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestChainStreamServerInterceptors_Multiple(t *testing.T) {
	name1 := "test-multi-stream-server-1"
	name2 := "test-multi-stream-server-2"

	callOrder := []string{}

	RegisterStreamServerIntBuilder(name1, func() StreamServerInterceptor {
		return func(srv interface{}, ss stream.ServerStream, info *StreamServerInfo, handler stream.StreamHandler) error {
			callOrder = append(callOrder, name1)
			return handler(srv, ss)
		}
	})

	RegisterStreamServerIntBuilder(name2, func() StreamServerInterceptor {
		return func(srv interface{}, ss stream.ServerStream, info *StreamServerInfo, handler stream.StreamHandler) error {
			callOrder = append(callOrder, name2)
			return handler(srv, ss)
		}
	})

	interceptor := ChainStreamServerInterceptors([]string{name1, name2})
	require.NotNil(t, interceptor)

	called := false
	handler := func(srv interface{}, ss stream.ServerStream) error {
		called = true
		return nil
	}

	info := &StreamServerInfo{FullMethod: "/test.Service/Method"}
	ss := &mockServerStream{ctx: context.Background()}

	err := interceptor("test-service", ss, info, handler)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, []string{name1, name2}, callOrder)
}

func TestInterceptorErrorHandling(t *testing.T) {
	name := "test-error-interceptor"
	RegisterUnaryClientIntBuilder(name, func(serviceName string) UnaryClientInterceptor {
		return func(ctx context.Context, method string, req, reply interface{}, invoker UnaryInvoker) error {
			return errors.New("interceptor error")
		}
	})

	interceptor := ChainUnaryClientInterceptors("test-service", []string{name})
	require.NotNil(t, interceptor)

	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}) error {
		called = true
		return nil
	}

	err := interceptor(context.Background(), "TestMethod", "req", "reply", invoker)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interceptor error")
	assert.False(t, called) // Invoker should not be called due to error
}

func TestUnaryServerInfo(t *testing.T) {
	info := &UnaryServerInfo{
		Server:     &struct{}{},
		FullMethod: "/test.Service/Method",
	}

	assert.NotNil(t, info.Server)
	assert.Equal(t, "/test.Service/Method", info.FullMethod)
}

func TestStreamServerInfo(t *testing.T) {
	info := &StreamServerInfo{
		FullMethod:     "/test.Service/Method",
		IsClientStream: true,
		IsServerStream: false,
	}

	assert.Equal(t, "/test.Service/Method", info.FullMethod)
	assert.True(t, info.IsClientStream)
	assert.False(t, info.IsServerStream)
}
