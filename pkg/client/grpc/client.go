package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getConfig(name string) *Config {
	c := &Config{}
	if err := config.Scan(fmt.Sprintf("yggdrasil.client.{%s}.grpc", name), c); err != nil {
		log.Fatalf("fault to get config, err: %s", err.Error())
		return nil
	}
	c.Name = name
	c.UnaryFilter = append(config.GetStringSlice("yggdrasil.grpc.unaryFilter"), c.UnaryFilter...)
	c.StreamFilter = append(config.GetStringSlice("yggdrasil.grpc.streamFilter"), c.StreamFilter...)
	return c
}

func DialByConfig(ctx context.Context, config *Config) *grpc.ClientConn {
	config.SetDefault()
	for _, name := range config.UnaryFilter {
		f, ok := unaryInterceptor[name]
		if !ok {
			log.Warnf("not found grpc unary interceptor, name: %s", name)
			continue
		}
		config.WithUnaryInterceptor(f(config.Name))
	}
	for _, name := range config.StreamFilter {
		f, ok := streamInterceptor[name]
		if !ok {
			log.Warnf("not found grpc stream interceptor, name: %s", name)
			continue
		}
		config.WithStreamInterceptor(f(config.Name))
	}
	config.WithDialOption(
		grpc.WithChainUnaryInterceptor(config.unaryInterceptors...),
		grpc.WithChainStreamInterceptor(config.streamInterceptors...),
	)

	if config.TLS != nil {
		tlsConfig, err := config.TLS.ServerTLSConfig()
		if err != nil {
			log.Fatalf("fault to get tls config, err: %s", err.Error())
		}
		config.WithDialOption(grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		config.WithDialOption(grpc.WithInsecure())
	}

	// 默认配置使用block
	if config.Block {
		if config.DialTimeout > time.Duration(0) {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, config.DialTimeout)
			defer cancel()
		}
		config.WithDialOption(grpc.WithBlock())
	}
	if config.KeepAlive != nil {
		config.WithDialOption(grpc.WithKeepaliveParams(*config.KeepAlive))
	}
	if config.Balancer != "" {
		config.WithDialOption(
			grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, config.Balancer)),
		)
	}
	cc, err := grpc.DialContext(ctx, config.Target, config.dialOptions...)
	if err != nil {
		if config.OnDialError == "panic" {
			log.Fatalf("dial grpc\terr: %s", err.Error())
		} else {
			log.Errorf("dial grpc\terr: %s", err.Error())
		}
	}
	return cc
}

func DialContext(ctx context.Context, ServerName string, opts ...grpc.DialOption) *grpc.ClientConn {
	return DialByConfig(ctx, getConfig(ServerName).WithDialOption(opts...))
}

func Dial(ServerName string, opts ...grpc.DialOption) *grpc.ClientConn {
	return DialByConfig(context.Background(), getConfig(ServerName).WithDialOption(opts...))
}
