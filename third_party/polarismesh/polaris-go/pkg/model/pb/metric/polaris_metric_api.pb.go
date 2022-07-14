// Code generated by protoc-gen-go.
// source: polaris_metric_api.proto
// DO NOT EDIT!

package metric

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for RateLimitGRPC service

type RateLimitGRPCClient interface {
	// 限流KEY初始化
	InitializeQuota(ctx context.Context, in *RateLimitRequest, opts ...grpc.CallOption) (*RateLimitResponse, error)
	// 获取限流配额
	AcquireQuota(ctx context.Context, opts ...grpc.CallOption) (RateLimitGRPC_AcquireQuotaClient, error)
}

type rateLimitGRPCClient struct {
	cc *grpc.ClientConn
}

func NewRateLimitGRPCClient(cc *grpc.ClientConn) RateLimitGRPCClient {
	return &rateLimitGRPCClient{cc}
}

func (c *rateLimitGRPCClient) InitializeQuota(ctx context.Context, in *RateLimitRequest, opts ...grpc.CallOption) (*RateLimitResponse, error) {
	out := new(RateLimitResponse)
	err := grpc.Invoke(ctx, "/v1.RateLimitGRPC/InitializeQuota", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rateLimitGRPCClient) AcquireQuota(ctx context.Context, opts ...grpc.CallOption) (RateLimitGRPC_AcquireQuotaClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_RateLimitGRPC_serviceDesc.Streams[0], c.cc, "/v1.RateLimitGRPC/AcquireQuota", opts...)
	if err != nil {
		return nil, err
	}
	x := &rateLimitGRPCAcquireQuotaClient{stream}
	return x, nil
}

type RateLimitGRPC_AcquireQuotaClient interface {
	Send(*RateLimitRequest) error
	Recv() (*RateLimitResponse, error)
	grpc.ClientStream
}

type rateLimitGRPCAcquireQuotaClient struct {
	grpc.ClientStream
}

func (x *rateLimitGRPCAcquireQuotaClient) Send(m *RateLimitRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *rateLimitGRPCAcquireQuotaClient) Recv() (*RateLimitResponse, error) {
	m := new(RateLimitResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for RateLimitGRPC service

type RateLimitGRPCServer interface {
	// 限流KEY初始化
	InitializeQuota(context.Context, *RateLimitRequest) (*RateLimitResponse, error)
	// 获取限流配额
	AcquireQuota(RateLimitGRPC_AcquireQuotaServer) error
}

func RegisterRateLimitGRPCServer(s *grpc.Server, srv RateLimitGRPCServer) {
	s.RegisterService(&_RateLimitGRPC_serviceDesc, srv)
}

func _RateLimitGRPC_InitializeQuota_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RateLimitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RateLimitGRPCServer).InitializeQuota(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.RateLimitGRPC/InitializeQuota",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RateLimitGRPCServer).InitializeQuota(ctx, req.(*RateLimitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RateLimitGRPC_AcquireQuota_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RateLimitGRPCServer).AcquireQuota(&rateLimitGRPCAcquireQuotaServer{stream})
}

type RateLimitGRPC_AcquireQuotaServer interface {
	Send(*RateLimitResponse) error
	Recv() (*RateLimitRequest, error)
	grpc.ServerStream
}

type rateLimitGRPCAcquireQuotaServer struct {
	grpc.ServerStream
}

func (x *rateLimitGRPCAcquireQuotaServer) Send(m *RateLimitResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *rateLimitGRPCAcquireQuotaServer) Recv() (*RateLimitRequest, error) {
	m := new(RateLimitRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _RateLimitGRPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "v1.RateLimitGRPC",
	HandlerType: (*RateLimitGRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InitializeQuota",
			Handler:    _RateLimitGRPC_InitializeQuota_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "AcquireQuota",
			Handler:       _RateLimitGRPC_AcquireQuota_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "polaris_metric_api.proto",
}

// Client API for MetricGRPC service

type MetricGRPCClient interface {
	// 初始化统计周期
	Init(ctx context.Context, in *MetricInitRequest, opts ...grpc.CallOption) (*MetricResponse, error)
	// 查询汇总统计数据
	Query(ctx context.Context, opts ...grpc.CallOption) (MetricGRPC_QueryClient, error)
	// 上报统计数据，并返回上报状态（成功or失败）
	Report(ctx context.Context, opts ...grpc.CallOption) (MetricGRPC_ReportClient, error)
}

type metricGRPCClient struct {
	cc *grpc.ClientConn
}

func NewMetricGRPCClient(cc *grpc.ClientConn) MetricGRPCClient {
	return &metricGRPCClient{cc}
}

func (c *metricGRPCClient) Init(ctx context.Context, in *MetricInitRequest, opts ...grpc.CallOption) (*MetricResponse, error) {
	out := new(MetricResponse)
	err := grpc.Invoke(ctx, "/v1.MetricGRPC/Init", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricGRPCClient) Query(ctx context.Context, opts ...grpc.CallOption) (MetricGRPC_QueryClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_MetricGRPC_serviceDesc.Streams[0], c.cc, "/v1.MetricGRPC/Query", opts...)
	if err != nil {
		return nil, err
	}
	x := &metricGRPCQueryClient{stream}
	return x, nil
}

type MetricGRPC_QueryClient interface {
	Send(*MetricQueryRequest) error
	Recv() (*MetricResponse, error)
	grpc.ClientStream
}

type metricGRPCQueryClient struct {
	grpc.ClientStream
}

func (x *metricGRPCQueryClient) Send(m *MetricQueryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *metricGRPCQueryClient) Recv() (*MetricResponse, error) {
	m := new(MetricResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *metricGRPCClient) Report(ctx context.Context, opts ...grpc.CallOption) (MetricGRPC_ReportClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_MetricGRPC_serviceDesc.Streams[1], c.cc, "/v1.MetricGRPC/Report", opts...)
	if err != nil {
		return nil, err
	}
	x := &metricGRPCReportClient{stream}
	return x, nil
}

type MetricGRPC_ReportClient interface {
	Send(*MetricRequest) error
	Recv() (*MetricResponse, error)
	grpc.ClientStream
}

type metricGRPCReportClient struct {
	grpc.ClientStream
}

func (x *metricGRPCReportClient) Send(m *MetricRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *metricGRPCReportClient) Recv() (*MetricResponse, error) {
	m := new(MetricResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for MetricGRPC service

type MetricGRPCServer interface {
	// 初始化统计周期
	Init(context.Context, *MetricInitRequest) (*MetricResponse, error)
	// 查询汇总统计数据
	Query(MetricGRPC_QueryServer) error
	// 上报统计数据，并返回上报状态（成功or失败）
	Report(MetricGRPC_ReportServer) error
}

func RegisterMetricGRPCServer(s *grpc.Server, srv MetricGRPCServer) {
	s.RegisterService(&_MetricGRPC_serviceDesc, srv)
}

func _MetricGRPC_Init_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricInitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricGRPCServer).Init(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v1.MetricGRPC/Init",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricGRPCServer).Init(ctx, req.(*MetricInitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricGRPC_Query_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MetricGRPCServer).Query(&metricGRPCQueryServer{stream})
}

type MetricGRPC_QueryServer interface {
	Send(*MetricResponse) error
	Recv() (*MetricQueryRequest, error)
	grpc.ServerStream
}

type metricGRPCQueryServer struct {
	grpc.ServerStream
}

func (x *metricGRPCQueryServer) Send(m *MetricResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *metricGRPCQueryServer) Recv() (*MetricQueryRequest, error) {
	m := new(MetricQueryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _MetricGRPC_Report_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MetricGRPCServer).Report(&metricGRPCReportServer{stream})
}

type MetricGRPC_ReportServer interface {
	Send(*MetricResponse) error
	Recv() (*MetricRequest, error)
	grpc.ServerStream
}

type metricGRPCReportServer struct {
	grpc.ServerStream
}

func (x *metricGRPCReportServer) Send(m *MetricResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *metricGRPCReportServer) Recv() (*MetricRequest, error) {
	m := new(MetricRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _MetricGRPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "v1.MetricGRPC",
	HandlerType: (*MetricGRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Init",
			Handler:    _MetricGRPC_Init_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Query",
			Handler:       _MetricGRPC_Query_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Report",
			Handler:       _MetricGRPC_Report_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "polaris_metric_api.proto",
}

func init() { proto.RegisterFile("polaris_metric_api.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 223 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x28, 0xc8, 0xcf, 0x49,
	0x2c, 0xca, 0x2c, 0x8e, 0xcf, 0x4d, 0x2d, 0x29, 0xca, 0x4c, 0x8e, 0x4f, 0x2c, 0xc8, 0xd4, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x83, 0x88, 0x48, 0x89, 0xc3, 0x54, 0x14, 0x25, 0x96, 0xa4,
	0xe6, 0x64, 0xe6, 0x66, 0x96, 0x40, 0x14, 0x48, 0x89, 0xa0, 0x6a, 0x85, 0x88, 0x1a, 0x2d, 0x61,
	0xe4, 0xe2, 0x0d, 0x4a, 0x2c, 0x49, 0xf5, 0x01, 0xa9, 0x74, 0x0f, 0x0a, 0x70, 0x16, 0xf2, 0xe0,
	0xe2, 0xf7, 0xcc, 0xcb, 0x2c, 0xc9, 0x4c, 0xcc, 0xc9, 0xac, 0x4a, 0x0d, 0x2c, 0xcd, 0x2f, 0x49,
	0x14, 0x92, 0xd0, 0x83, 0xea, 0x81, 0xab, 0x0c, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0x91, 0x92,
	0xc4, 0x22, 0x53, 0x5c, 0x90, 0x9f, 0x57, 0x9c, 0xaa, 0xc4, 0x20, 0xe4, 0xc9, 0xc5, 0xe3, 0x98,
	0x5c, 0x58, 0x9a, 0x59, 0x44, 0x91, 0x31, 0x1a, 0x8c, 0x06, 0x8c, 0x46, 0xa7, 0x19, 0xb9, 0xb8,
	0x7c, 0xc1, 0x4a, 0xc0, 0x6e, 0xb4, 0xe6, 0x62, 0x01, 0xb9, 0x51, 0x08, 0xae, 0x0f, 0x22, 0x07,
	0x12, 0x83, 0x19, 0x29, 0x86, 0x2a, 0x85, 0xe4, 0x2c, 0x47, 0x2e, 0xd6, 0xc0, 0xd2, 0xd4, 0xa2,
	0x4a, 0x21, 0x29, 0x54, 0x25, 0x60, 0x41, 0x82, 0xda, 0x41, 0xce, 0x11, 0xb2, 0xe5, 0x62, 0x0b,
	0x4a, 0x2d, 0xc8, 0x2f, 0x2a, 0x11, 0x12, 0x45, 0x57, 0x47, 0x84, 0xf6, 0x24, 0x36, 0x70, 0xd8,
	0x1b, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x5a, 0x90, 0x7a, 0x84, 0xce, 0x01, 0x00, 0x00,
}
