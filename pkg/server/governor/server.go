package governor

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
)

func init() {
	governor.Init()
	server.RegisterConstructor(constructor)
}

// Server ...
type Server struct {
	*http.Server
	listener net.Listener
	*Config
	info types.ServerInfo
}

func newServer(config *Config) *Server {
	var listener, err = net.Listen("tcp4", config.Address())
	if err != nil {
		log.Fatalf("governor start error: %s", err.Error())
	}
	config.Host, config.Port = xnet.GetHostAndPortByAddr(listener.Addr())
	return &Server{
		Server: &http.Server{
			Addr:    config.Address(),
			Handler: governor.DefaultServeMux,
		},
		listener: listener,
		Config:   config,
		info: server.NewInfo(
			"http",
			types.ServerKindGovernor,
			fmt.Sprintf("%s:%d", config.Host, config.Port),
			map[string]string{},
		),
	}
}

// Serve ..
func (s *Server) Serve() error {
	err := s.Server.Serve(s.listener)
	if err == http.ErrServerClosed {
		return nil
	}
	return err

}

// Shutdown ..
func (s *Server) Stop() error {
	return s.Server.Shutdown(context.TODO())
}

// Info ..
func (s *Server) Info() types.ServerInfo {
	return s.info
}

func constructor() types.Server {
	cfg := &Config{}
	if err := config.Scan("yggdrasil.server.governor", cfg); err != nil {
		log.Fatalf("fault to scan grpc server config, err: %s", err.Error())
		return nil
	}
	return cfg.Build()
}
