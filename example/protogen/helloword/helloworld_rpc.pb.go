// Code generated by protoc-gen-yggdrasil-rpc. DO NOT EDIT.

package helloword

import (
	context "context"
	client "github.com/imkuqin-zw/yggdrasil/pkg/client"
	interceptor "github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	metadata "github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	server "github.com/imkuqin-zw/yggdrasil/pkg/server"
	status "github.com/imkuqin-zw/yggdrasil/pkg/status"
	stream "github.com/imkuqin-zw/yggdrasil/pkg/stream"
	code "google.golang.org/genproto/googleapis/rpc/code"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the yggdrasil package it is being compiled against.
var _ = new(metadata.MD)

type GreeterClient interface {
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
	SayError(context.Context, *HelloRequest) (*HelloReply, error)
	SayHelloStream(context.Context) (GreeterSayHelloStreamClient, error)
	SayHelloClientStream(context.Context) (GreeterSayHelloClientStreamClient, error)
	SayHelloServerStream(context.Context, *HelloRequest) (GreeterSayHelloServerStreamClient, error)
}

type GreeterSayHelloStreamClient interface {
	Send(*HelloRequest) error
	Recv() (*HelloReply, error)
	stream.ClientStream
}

type GreeterSayHelloClientStreamClient interface {
	Send(*HelloRequest) error
	stream.ClientStream
}

type GreeterSayHelloServerStreamClient interface {
	CloseAndRecv() (*HelloReply, error)
	Recv() (*HelloReply, error)
	stream.ClientStream
}

type greeterClient struct {
	cc client.Client
}

func NewGreeterClient(cc client.Client) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	out := new(HelloReply)
	err := c.cc.Invoke(ctx, "/yggdrasil.example.proto.helloword.Greeter/SayHello", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greeterClient) SayError(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	out := new(HelloReply)
	err := c.cc.Invoke(ctx, "/yggdrasil.example.proto.helloword.Greeter/SayError", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greeterClient) SayHelloStream(ctx context.Context) (GreeterSayHelloStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &GreeterServiceDesc.Streams[0], "/yggdrasil.example.proto.helloword.Greeter/SayHelloStream")
	if err != nil {
		return nil, err
	}
	x := &greeterSayHelloStreamClient{stream}
	return x, nil
}

type greeterSayHelloStreamClient struct {
	stream.ClientStream
}

func (x *greeterSayHelloStreamClient) Header() (metadata.MD, error) {
	v, err := x.ClientStream.Header()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (x *greeterSayHelloStreamClient) Trailer() metadata.MD {
	return x.ClientStream.Trailer()
}

func (x *greeterSayHelloStreamClient) Send(m *HelloRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *greeterSayHelloStreamClient) Recv() (*HelloReply, error) {
	m := new(HelloReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
func (c *greeterClient) SayHelloClientStream(ctx context.Context) (GreeterSayHelloClientStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &GreeterServiceDesc.Streams[1], "/yggdrasil.example.proto.helloword.Greeter/SayHelloClientStream")
	if err != nil {
		return nil, err
	}
	x := &greeterSayHelloClientStreamClient{stream}
	return x, nil
}

type greeterSayHelloClientStreamClient struct {
	stream.ClientStream
}

func (x *greeterSayHelloClientStreamClient) Header() (metadata.MD, error) {
	v, err := x.ClientStream.Header()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (x *greeterSayHelloClientStreamClient) Trailer() metadata.MD {
	return x.ClientStream.Trailer()
}

func (x *greeterSayHelloClientStreamClient) Send(m *HelloRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (c *greeterClient) SayHelloServerStream(ctx context.Context, in *HelloRequest) (GreeterSayHelloServerStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &GreeterServiceDesc.Streams[2], "/yggdrasil.example.proto.helloword.Greeter/SayHelloServerStream")
	if err != nil {
		return nil, err
	}
	x := &greeterSayHelloServerStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type greeterSayHelloServerStreamClient struct {
	stream.ClientStream
}

func (x *greeterSayHelloServerStreamClient) CloseAndRecv() (*HelloReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(HelloReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *greeterSayHelloServerStreamClient) Recv() (*HelloReply, error) {
	m := new(HelloReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *greeterSayHelloServerStreamClient) Header() (metadata.MD, error) {
	v, err := x.ClientStream.Header()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (x *greeterSayHelloServerStreamClient) Trailer() metadata.MD {
	return x.ClientStream.Trailer()
}

func (x *greeterSayHelloServerStreamClient) Send(m *HelloRequest) error {
	return x.ClientStream.SendMsg(m)
}

func _Greeter_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, unaryInt interceptor.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if unaryInt == nil {
		return srv.(GreeterServer).SayHello(ctx, in)
	}
	info := &interceptor.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yggdrasil.example.proto.helloword.Greeter/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).SayHello(ctx, req.(*HelloRequest))
	}
	return unaryInt(ctx, in, info, handler)
}

func _Greeter_SayError_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, unaryInt interceptor.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if unaryInt == nil {
		return srv.(GreeterServer).SayError(ctx, in)
	}
	info := &interceptor.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yggdrasil.example.proto.helloword.Greeter/SayError",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).SayError(ctx, req.(*HelloRequest))
	}
	return unaryInt(ctx, in, info, handler)
}

func _Greeter_SayHelloStream_Handler(srv interface{}, stream stream.ServerStream) error {
	return srv.(GreeterServer).SayHelloStream(&greeterSayHelloStreamServer{stream})
}

type greeterSayHelloStreamServer struct {
	stream.ServerStream
}

func (x *greeterSayHelloStreamServer) SetHeader(md metadata.MD) error {
	return x.ServerStream.SetHeader(md)
}

func (x *greeterSayHelloStreamServer) SendHeader(md metadata.MD) error {
	return x.ServerStream.SendHeader(md)
}

func (x *greeterSayHelloStreamServer) SetTrailer(md metadata.MD) {
	x.ServerStream.SetTrailer(md)
}

func (x *greeterSayHelloStreamServer) Context() context.Context {
	return x.ServerStream.Context()
}

func (x *greeterSayHelloStreamServer) Send(m *HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greeterSayHelloStreamServer) Recv() (*HelloRequest, error) {
	m := new(HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Greeter_SayHelloClientStream_Handler(srv interface{}, stream stream.ServerStream) error {
	return srv.(GreeterServer).SayHelloClientStream(&greeterSayHelloClientStreamServer{stream})
}

type greeterSayHelloClientStreamServer struct {
	stream.ServerStream
}

func (x *greeterSayHelloClientStreamServer) SetHeader(md metadata.MD) error {
	return x.ServerStream.SetHeader(md)
}

func (x *greeterSayHelloClientStreamServer) SendHeader(md metadata.MD) error {
	return x.ServerStream.SendHeader(md)
}

func (x *greeterSayHelloClientStreamServer) SetTrailer(md metadata.MD) {
	x.ServerStream.SetTrailer(md)
}

func (x *greeterSayHelloClientStreamServer) Context() context.Context {
	return x.ServerStream.Context()
}

func (x *greeterSayHelloClientStreamServer) SendAndClose(m *HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greeterSayHelloClientStreamServer) Recv() (*HelloRequest, error) {
	m := new(HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Greeter_SayHelloServerStream_Handler(srv interface{}, stream stream.ServerStream) error {
	m := new(HelloRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GreeterServer).SayHelloServerStream(m, &greeterSayHelloServerStreamServer{stream})
}

type greeterSayHelloServerStreamServer struct {
	stream.ServerStream
}

func (x *greeterSayHelloServerStreamServer) SetHeader(md metadata.MD) error {
	return x.ServerStream.SetHeader(md)
}

func (x *greeterSayHelloServerStreamServer) SendHeader(md metadata.MD) error {
	return x.ServerStream.SendHeader(md)
}

func (x *greeterSayHelloServerStreamServer) SetTrailer(md metadata.MD) {
	x.ServerStream.SetTrailer(md)
}

func (x *greeterSayHelloServerStreamServer) Context() context.Context {
	return x.ServerStream.Context()
}

func (x *greeterSayHelloServerStreamServer) Send(m *HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

type GreeterServer interface {
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
	SayError(context.Context, *HelloRequest) (*HelloReply, error)
	SayHelloStream(GreeterSayHelloStreamServer) error
	SayHelloClientStream(GreeterSayHelloClientStreamServer) error
	SayHelloServerStream(*HelloRequest, GreeterSayHelloServerStreamServer) error
	UnsafeGreeterServer
}

type GreeterSayHelloStreamServer interface {
	Send(*HelloReply) error
	Recv() (*HelloRequest, error)
	stream.ServerStream
}

type GreeterSayHelloClientStreamServer interface {
	SendAndClose(*HelloReply) error
	Recv() (*HelloRequest, error)
	stream.ServerStream
}

type GreeterSayHelloServerStreamServer interface {
	Send(*HelloReply) error
	stream.ServerStream
}

type UnsafeGreeterServer interface {
	mustEmbedUnimplementedGreeterServer()
}

// UnimplementedGreeterServer must be embedded to have forward compatible implementations.
type UnimplementedGreeterServer struct {
}

func (UnimplementedGreeterServer) SayHello(context.Context, *HelloRequest) (*HelloReply, error) {
	return nil, status.Errorf(code.Code_UNIMPLEMENTED, "method SayHello not implemented")
}

func (UnimplementedGreeterServer) SayError(context.Context, *HelloRequest) (*HelloReply, error) {
	return nil, status.Errorf(code.Code_UNIMPLEMENTED, "method SayError not implemented")
}

func (UnimplementedGreeterServer) SayHelloStream(GreeterSayHelloStreamServer) error {
	return status.Errorf(code.Code_UNIMPLEMENTED, "method SayHelloStream not implemented")
}

func (UnimplementedGreeterServer) SayHelloClientStream(GreeterSayHelloClientStreamServer) error {
	return status.Errorf(code.Code_UNIMPLEMENTED, "method SayHelloClientStream not implemented")
}

func (UnimplementedGreeterServer) SayHelloServerStream(*HelloRequest, GreeterSayHelloServerStreamServer) error {
	return status.Errorf(code.Code_UNIMPLEMENTED, "method SayHelloServerStream not implemented")
}

func (UnimplementedGreeterServer) mustEmbedUnimplementedGreeterServer() {}

var GreeterServiceDesc = server.ServiceDesc{
	ServiceName: "yggdrasil.example.proto.helloword.Greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []server.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Greeter_SayHello_Handler,
		},
		{
			MethodName: "SayError",
			Handler:    _Greeter_SayError_Handler,
		},
	},
	Streams: []stream.StreamDesc{
		{
			StreamName:    "SayHelloStream",
			Handler:       _Greeter_SayHelloStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "SayHelloClientStream",
			Handler:       _Greeter_SayHelloClientStream_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "SayHelloServerStream",
			Handler:       _Greeter_SayHelloServerStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "helloword/helloworld.proto",
}
