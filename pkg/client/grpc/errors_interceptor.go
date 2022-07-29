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

	"github.com/imkuqin-zw/yggdrasil/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func init() {
	RegisterUnaryInterceptor("error", func(string) grpc.UnaryClientInterceptor { return ErrorUnaryClientInterceptor })
	RegisterStreamInterceptor("error", func(string) grpc.StreamClientInterceptor { return ErrorStreamClientInterceptor })
}

func ErrorUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		return errors.FromProto(status.Convert(err).Proto())
	}
	return err
}

func ErrorStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	cs, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return cs, errors.FromProto(status.Convert(err).Proto())
	}
	return cs, err
}
