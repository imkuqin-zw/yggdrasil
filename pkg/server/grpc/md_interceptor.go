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
	ctx = md.WithStreamContext(ctx)
	if meta, ok := metadata.FromIncomingContext(ctx); ok {
		if m, ok := md.FromInContext(ctx); ok {
			ctx = md.WithInContext(ctx, md.Join(m, md.MD(meta)))
		} else {
			ctx = md.WithInContext(ctx, md.MD(meta))
		}
	}
	res, resErr := handler(ctx, req)
	if header, ok := md.FromHeaderCtx(ctx); ok {
		if err := grpc.SetHeader(ctx, metadata.MD(header)); err != nil {
			return nil, err
		}
	}
	if trailer, ok := md.FromTrailerCtx(ctx); ok {
		if err := grpc.SetTrailer(ctx, metadata.MD(trailer)); err != nil {
			return nil, err
		}
	}
	return res, resErr
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
	var ctx = md.WithStreamContext(stream.Context())
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
