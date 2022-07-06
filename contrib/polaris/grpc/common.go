package grpc

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&resolverBuilder{})
	balancer.Register(&balancerBuilder{})
	// grpc2.RegisterUnaryInterceptor(newLimitUnaryInterceptor())
}

const scheme = "polaris"
const keyDialOptions = "options"

type dialOpts struct {
	Namespace   string
	DstMetadata map[string]string
	SrcMetadata map[string]string
	SrcService  string
	// 可选，规则路由Meta匹配前缀，用于过滤作为路由规则的gRPC Header
	HeaderPrefix []string
}
