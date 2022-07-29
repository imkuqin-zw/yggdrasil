package grpc

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
	"google.golang.org/grpc"
)

func init() {
	server.RegisterConstructor(constructor)
}

type service struct {
	desc *grpc.ServiceDesc
	impl interface{}
}

var services = map[string]service{}

func RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	services[desc.ServiceName] = service{
		desc: desc,
		impl: impl,
	}
}

type grpcServer struct {
	cfg    *Config
	ln     net.Listener
	server *grpc.Server
	info   types.ServerInfo
}

func newServer(cfg *Config) *grpcServer {
	cfg.serverOptions = append(cfg.serverOptions,
		grpc.ChainStreamInterceptor(cfg.streamInterceptors...),
		grpc.ChainUnaryInterceptor(cfg.unaryInterceptors...),
	)
	var (
		ln  net.Listener
		err error
	)
	if cfg.TLS != nil {
		tlsConfig, err := cfg.TLS.ServerTLSConfig()
		if err != nil {
			log.Fatalf("fault to get tls config, err: %s", err.Error())
		}
		ln, err = tls.Listen(cfg.Network, cfg.Address(), tlsConfig)
	} else {
		ln, err = net.Listen(cfg.Network, cfg.Address())
	}
	if err != nil {
		log.Fatalf("fault to get listener, err: %s", err.Error())
		return nil
	}
	cfg.Host, cfg.Port = xnet.GetHostAndPortByAddr(ln.Addr())
	return &grpcServer{
		cfg: cfg,
		ln:  ln,
		info: server.NewInfo(
			"grpc",
			types.ServerKindRpc,
			fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			map[string]string{},
		)}
}

func (s *grpcServer) Serve() error {
	s.server = grpc.NewServer(s.cfg.serverOptions...)
	for _, service := range services {
		s.server.RegisterService(service.desc, service.impl)
		methods := make([]string, len(service.desc.Methods))
		for i, method := range service.desc.Methods {
			methods[i] = method.MethodName
		}
		governor.RegisterService(service.desc.ServiceName, methods)
	}
	return s.server.Serve(s.ln)
}

func (s *grpcServer) Stop() error {
	s.server.GracefulStop()
	return nil
}

func (s *grpcServer) Info() types.ServerInfo {
	return s.info
}

func constructor() types.Server {
	cfg := &Config{}
	if err := config.Scan("yggdrasil.server.grpc", cfg); err != nil {
		log.Fatalf("fault to scan grpc server config, err: %s", err.Error())
		return nil
	}
	return cfg.Build()
}
