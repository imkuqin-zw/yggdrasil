package grpc

import (
	"google.golang.org/grpc"
)

var unaryInterceptor = map[string]func(serverName string) grpc.UnaryClientInterceptor{}
var streamInterceptor = map[string]func(serverName string) grpc.StreamClientInterceptor{}

func RegisterUnaryInterceptor(name string, interceptor func(serverName string) grpc.UnaryClientInterceptor) {
	unaryInterceptor[name] = interceptor
}
func RegisterStreamInterceptor(name string, interceptor func(serverName string) grpc.StreamClientInterceptor) {
	streamInterceptor[name] = interceptor
}
