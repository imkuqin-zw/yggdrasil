package grpc

import (
	"google.golang.org/grpc"
)

var unaryInterceptor = map[string]func() grpc.UnaryServerInterceptor{}
var streamInterceptor = map[string]func() grpc.StreamServerInterceptor{}
var serverOptions map[string]func() grpc.ServerOption

func RegisterUnaryInterceptor(name string, interceptor func() grpc.UnaryServerInterceptor) {
	unaryInterceptor[name] = interceptor
}

func RegisterStreamInterceptor(name string, interceptor func() grpc.StreamServerInterceptor) {
	streamInterceptor[name] = interceptor
}

func RegisterServerOptions(name string, opt func() grpc.ServerOption) {
	serverOptions[name] = opt
}
