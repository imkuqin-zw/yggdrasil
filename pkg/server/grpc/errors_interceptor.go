package grpc

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/errors"
	"google.golang.org/grpc"
)

func init() {
	RegisterUnaryInterceptor("error", func() grpc.UnaryServerInterceptor { return ErrorUnaryServerInterceptor })
	RegisterStreamInterceptor("error", func() grpc.StreamServerInterceptor { return ErrorStreamServerInterceptor })
}

func ErrorUnaryServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, errors.FromError(err)
	}
	return resp, nil
}

func ErrorStreamServerInterceptor(srv interface{}, stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if err := handler(srv, stream); err != nil {
		return errors.FromError(err)
	}
	return nil
}
