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

func newServer(cfg *Config) *Server {
	var listener, err = net.Listen("tcp4", cfg.Address())
	if err != nil {
		log.Fatalf("governor start error: %s", err.Error())
	}
	cfg.Host, cfg.Port = xnet.GetHostAndPortByAddr(listener.Addr())
	return &Server{
		Server: &http.Server{
			Addr:    cfg.Address(),
			Handler: governor.DefaultServeMux,
		},
		listener: listener,
		Config:   cfg,
		info: server.NewInfo(
			"http",
			types.ServerKindGovernor,
			fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
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
