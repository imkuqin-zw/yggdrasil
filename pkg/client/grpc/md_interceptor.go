package grpc

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/md"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	RegisterUnaryInterceptor("metadata", func(string) grpc.UnaryClientInterceptor { return MdUnaryClientInterceptor })
	RegisterStreamInterceptor("metadata", func(string) grpc.StreamClientInterceptor { return MdStreamClientInterceptor })
}

func MdUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	meta, ok := md.FromOutContext(ctx)
	if ok {
		if m, ok := metadata.FromOutgoingContext(ctx); ok {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Join(m, metadata.MD(meta)))
		} else {
			ctx = metadata.NewOutgoingContext(ctx, metadata.MD(meta))
		}
	}

	var (
		header  *metadata.MD
		trailer *metadata.MD
	)
	if _, ok := md.FromHeaderCtx(ctx); ok {
		header = &metadata.MD{}
		opts = append(opts, grpc.HeaderCallOption{HeaderAddr: header})
	}
	if _, ok := md.FromTrailerCtx(ctx); ok {
		trailer = &metadata.MD{}
		opts = append(opts, grpc.TrailerCallOption{TrailerAddr: trailer})
	}
	err := invoker(ctx, method, req, reply, cc, opts...)
	if header != nil {
		_ = md.SetHeader(ctx, md.MD(*header))
	}
	if trailer != nil {
		_ = md.SetTrailer(ctx, md.MD(*trailer))
	}
	return err
}

func MdStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	meta, ok := md.FromOutContext(ctx)
	if ok {
		if m, ok := metadata.FromOutgoingContext(ctx); ok {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Join(m, metadata.MD(meta)))
		} else {
			ctx = metadata.NewOutgoingContext(ctx, metadata.MD(meta))
		}
	}
	return streamer(ctx, desc, cc, method, opts...)
}
