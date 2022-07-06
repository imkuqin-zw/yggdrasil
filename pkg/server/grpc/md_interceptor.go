package grpc

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/md"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	RegisterUnaryInterceptor("metadata", func() grpc.UnaryServerInterceptor { return MdUnaryServerInterceptor })
	RegisterStreamInterceptor("metadata", func() grpc.StreamServerInterceptor { return MdStreamServerInterceptor })
}

func MdUnaryServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		if m, ok := md.FromInContext(ctx); ok {
			ctx = md.WithInContext(ctx, md.Join(m, md.MD(meta)))
		} else {
			ctx = md.WithInContext(ctx, md.MD(meta))
		}
	}
	return handler(ctx, req)
}

type mdServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *mdServerStream) Context() context.Context {
	return ss.ctx
}

func MdStreamServerInterceptor(srv interface{}, stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	var ctx = stream.Context()
	if meta, ok := metadata.FromIncomingContext(stream.Context()); ok {
		if m, ok := md.FromInContext(ctx); ok {
			ctx = md.WithInContext(ctx, md.Join(m, md.MD(meta)))
		} else {
			ctx = md.WithInContext(ctx, md.MD(meta))
		}
	}
	return handler(srv, &mdServerStream{
		ServerStream: stream,
		ctx:          ctx,
	})
}
