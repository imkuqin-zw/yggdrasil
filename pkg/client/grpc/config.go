package grpc

import (
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xtls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultSlowThreshold = time.Millisecond * 500
	defaultReadTimeout   = time.Second * 2
	defaultDialTimeout   = time.Second * 3
)

// Config ...
type Config struct {
	Name        string // config's name
	Balancer    string
	Target      string
	Block       bool
	DialTimeout time.Duration
	ReadTimeout time.Duration
	OnDialError string // panic | error
	KeepAlive   *keepalive.ClientParameters
	TLS         *xtls.SSLConfig

	UnaryFilter  []string
	StreamFilter []string

	dialOptions        []grpc.DialOption
	streamInterceptors []grpc.StreamClientInterceptor
	unaryInterceptors  []grpc.UnaryClientInterceptor
}

// WithDialOption ...
func (c *Config) WithDialOption(opts ...grpc.DialOption) *Config {
	c.dialOptions = append(c.dialOptions, opts...)
	return c
}

// WithStreamInterceptor inject stream interceptors to grpc client option
func (c *Config) WithStreamInterceptor(interceptors ...grpc.StreamClientInterceptor) *Config {
	c.streamInterceptors = append(c.streamInterceptors, interceptors...)
	return c
}

// WithUnaryInterceptor inject unary interceptors to grpc client option
func (c *Config) WithUnaryInterceptor(interceptors ...grpc.UnaryClientInterceptor) *Config {
	c.unaryInterceptors = append(c.unaryInterceptors, interceptors...)
	return c
}

func (c *Config) SetDefault() {
	if c.DialTimeout == 0 {
		c.DialTimeout = defaultDialTimeout
	}

	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultReadTimeout
	}

	if c.Balancer == "" {
		c.Balancer = roundrobin.Name
	}
	c.UnaryFilter = xarray.RemoveReplaceStrings(append([]string{"error", "metadata", "log"}, c.UnaryFilter...))
	c.StreamFilter = xarray.RemoveReplaceStrings(append([]string{"error", "metadata", "log"}, c.StreamFilter...))
}
