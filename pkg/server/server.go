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

package server

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/genproto/googleapis/rpc/code"
)

const (
	serverStateInit = iota
	serverStateRunning
	serverStateClosing
)

type serverInfo struct {
	scheme   string
	address  string
	metadata map[string]string
}

func (si *serverInfo) Address() string {
	return si.address
}

func (si *serverInfo) Metadata() map[string]string {
	return si.metadata
}

func (si *serverInfo) Scheme() string {
	return si.scheme
}

func NewInfo(scheme string, address string, metadata map[string]string) Endpoint {
	return &serverInfo{scheme: scheme, address: address, metadata: metadata}
}

var (
	s    *server
	once sync.Once
)

type server struct {
	mu                sync.RWMutex
	services          map[string]*ServiceInfo // service name -> service serverInfo
	unaryInterceptor  interceptor.UnaryServerInterceptor
	streamInterceptor interceptor.StreamServerInterceptor
	servers           []remote.Server
	state             int
}

func GetServer() Server {
	once.Do(func() {
		s = &server{
			services: map[string]*ServiceInfo{},
		}
		s.initInterceptor()
		s.initRemoteServer()
	})
	return s
}

// RegisterService registers a service and its implementation to the gRPC
// server. It is called from the IDL generated code. This must be called before
// invoking Serve. If ss is non-nil (for legacy code), its type is checked to
// ensure it implements sd.HandlerType.
func (s *server) RegisterService(sd *ServiceDesc, ss interface{}) {
	if ss != nil {
		ht := reflect.TypeOf(sd.HandlerType).Elem()
		st := reflect.TypeOf(ss)
		if !st.Implements(ht) {
			logger.Fatalf("Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
		}
	}
	s.register(sd, ss)
}

func (s *server) Stop() error {
	s.mu.Lock()
	if s.state == serverStateInit {
		s.state = serverStateClosing
		s.mu.Unlock()
		return nil
	}
	if s.state == serverStateClosing {
		s.mu.Unlock()
		return nil
	}
	s.state = serverStateClosing
	s.mu.Unlock()
	errs := make([]error, 0)
	for _, item := range s.servers {
		if err := item.Stop(); err != nil {
			errs = append(errs, err)
			logger.ErrorFiled("fault to stop server",
				logger.String("protocol", item.Info().Protocol), logger.Err(err))
		}
	}
	return multierr.Combine(errs...)
}

func (s *server) Serve() (<-chan struct{}, <-chan struct{}, <-chan error) {
	finishCh := make(chan error)
	s.mu.Lock()
	if s.state == serverStateClosing {
		s.mu.Unlock()
		finishCh <- errors.New("server stopped")
		return nil, nil, finishCh
	}
	if s.state == serverStateRunning {
		s.mu.Unlock()
		finishCh <- errors.New("server already serve")
		return nil, nil, finishCh
	}
	s.state = serverStateRunning
	s.mu.Unlock()
	var (
		servers      = s.servers
		initFinishCh = make(chan struct{})
		initNum      atomic.Int32
		serviceNum   = int32(len(services))
	)

	g, ctx := errgroup.WithContext(context.Background())
	go func() {
		defer close(finishCh)
		for _, item := range servers {
			svr := item
			g.Go(func() error {
				ch, err := svr.Serve()
				if err != nil {
					logger.ErrorFiled("the server was ended forcefully",
						logger.String("protocol", svr.Info().Protocol), logger.Err(err))
					return err
				}
				info := svr.Info()
				logger.InfoFiled("server start", logger.String("endpoint",
					fmt.Sprintf("%s://%s", info.Protocol, info.Address)))
				if initNum.Add(1) == serviceNum {
					close(initFinishCh)
				}
				err, _ = <-ch
				return err
			})
		}
		finishCh <- g.Wait()
	}()

	return ctx.Done(), initFinishCh, finishCh
}

func (s *server) Endpoints() []Endpoint {
	endpoints := make([]Endpoint, len(s.servers))
	for i, item := range s.servers {
		e := item.Info()
		endpoints[i] = NewInfo(e.Protocol, e.Address, e.Attr)
	}
	return endpoints
}

func (s *server) initInterceptor() {
	if val := config.Get(config.KeyIntUnaryServe).String(); val != "" {
		unaryIntNames := xarray.RemoveReplaceStrings(strings.Split(val, ","))
		s.unaryInterceptor = interceptor.ChainUnaryServerInterceptors(unaryIntNames)
	}
	if val := config.Get(config.KeyIntStreamServer).String(); val != "" {
		streamIntNames := xarray.RemoveReplaceStrings(strings.Split(val, ","))
		s.streamInterceptor = interceptor.ChainStreamServerInterceptors(streamIntNames)
	}
}

func (s *server) initRemoteServer() {
	protocols := config.Get(config.KeyServerProtocol).StringSlice()
	if len(protocols) == 0 {
		logger.Fatal(errors.New("server protocols can not be empty"))
	}
	for _, protocol := range protocols {
		builder := remote.GetServerBuilder(protocol)
		if builder == nil {
			logger.FatalFiled("not found server builder",
				logger.String("protocol", protocol))
		}
		svr, err := builder(s.handle)
		if err != nil {
			logger.FatalFiled("fault to new remote server",
				logger.String("protocol", protocol),
				logger.Err(err))
		}
		s.servers = append(s.servers, svr)
	}

}

func (s *server) register(sd *ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[sd.ServiceName]; ok {
		logger.Fatalf("Server.RegisterService found duplicate service registration for %q", sd.ServiceName)
	}
	registerService(sd)
	info := &ServiceInfo{
		ServiceImpl: ss,
		Methods:     make(map[string]*MethodDesc),
		Streams:     make(map[string]*stream.StreamDesc),
		Metadata:    sd.Metadata,
	}
	for i := range sd.Methods {
		d := &sd.Methods[i]
		info.Methods[d.MethodName] = d
	}
	for i := range sd.Streams {
		d := &sd.Streams[i]
		info.Streams[d.StreamName] = d
	}
	s.services[sd.ServiceName] = info
}

func (s *server) handle(ctx context.Context, sm string, ss stream.ServerStream) (interface{}, bool, error) {
	if sm != "" && sm[0] == '/' {
		sm = sm[1:]
	}
	pos := strings.LastIndex(sm, "/")
	if pos == -1 {
		errDesc := fmt.Sprintf("malformed method name: %q", sm)
		return nil, false, status.Errorf(code.Code_UNIMPLEMENTED, errDesc)
	}
	service := sm[:pos]
	method := sm[pos+1:]
	srv, knownService := s.services[service]
	if knownService {
		if md, ok := srv.Methods[method]; ok {
			reply, err := md.Handler(srv.ServiceImpl, ctx, ss.RecvMsg, s.unaryInterceptor)
			return reply, false, err
		}
		if sd, ok := srv.Streams[method]; ok {
			si := &interceptor.StreamServerInfo{
				FullMethod:     sm,
				IsClientStream: sd.ClientStreams,
				IsServerStream: sd.ServerStreams,
			}
			err := s.streamInterceptor(srv.ServiceImpl, ss, si, sd.Handler)
			return nil, true, err
		}
	}
	var errDesc string
	if !knownService {
		errDesc = fmt.Sprintf("unknown service %v", service)
	} else {
		errDesc = fmt.Sprintf("unknown method %v for service %v", method, service)
	}
	return nil, false, status.Errorf(code.Code_UNIMPLEMENTED, errDesc)
}
