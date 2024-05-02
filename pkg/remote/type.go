package remote

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
)

//type StatsFunc func(bool, bool) (func(), func(error))

//type MethodHandle func(sm string, ss stream.ServerStream, sf StatsFunc) (any, bool, error)

type MethodHandle func(ServerStream)

type Server interface {
	Start() error
	Handle() error
	Stop() error
	Info() ServerInfo
}

type ServerStream interface {
	stream.ServerStream
	Method() string
	Start(isClientStream, isServerStream bool) error
	Finish(any, error)
}

type Client interface {
	NewStream(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error)
	Close() error
	Scheme() string
}
