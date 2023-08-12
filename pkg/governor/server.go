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
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
)

func init() {
	Init()
}

type ServerInfo struct {
	Address string
	Scheme  string
	Attr    map[string]string
}

// Server ...
type Server struct {
	*http.Server
	listener net.Listener
	*Config
	info ServerInfo
}

func NewServer() *Server {
	cfg := &Config{}
	if err := config.Get(config.KeyGovernor).Scan(cfg); err != nil {
		logger.FatalField("fault to get governor config", logger.Err(err))
	}
	cfg.SetDefault()
	var listener, err = net.Listen("tcp4", cfg.Address())
	if err != nil {
		log.Fatalf("governor start reason: %s", err.Error())
	}
	cfg.Host, cfg.Port = xnet.GetHostAndPortByAddr(listener.Addr())
	return &Server{
		Server: &http.Server{
			Addr:    cfg.Address(),
			Handler: DefaultServeMux,
		},
		listener: listener,
		Config:   cfg,
		info: ServerInfo{
			Address: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Scheme:  "http",
			Attr:    map[string]string{},
		},
	}
}

// Serve ..
func (s *Server) Serve() error {
	info := s.Info()
	logger.InfoField("governor start", logger.String("endpoint", fmt.Sprintf("%s://%s", info.Scheme, info.Address)))
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
func (s *Server) Info() ServerInfo {
	return s.info
}
