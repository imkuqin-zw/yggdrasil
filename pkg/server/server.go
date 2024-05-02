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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote"
	"github.com/imkuqin-zw/yggdrasil/pkg/stats"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"go.uber.org/multierr"
	"google.golang.org/genproto/googleapis/rpc/code"
)

const (
	serverStateInit = iota
	serverStateRunning
	serverStateClosing
)

var (
	svr  *server
	once sync.Once
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

type server struct {
	mu                sync.RWMutex
	services          map[string]*ServiceInfo // service name -> service serverInfo
	servicesDesc      map[string][]methodInfo
	unaryInterceptor  interceptor.UnaryServerInterceptor
	streamInterceptor interceptor.StreamServerInterceptor
	servers           []remote.Server
	state             int
	serverWG          sync.WaitGroup
	stats             stats.Handler
}

func NetServer() Server {
	svr = &server{
		services:     map[string]*ServiceInfo{},
		servicesDesc: map[string][]methodInfo{},
		stats:        stats.GetServerHandler(),
	}
	svr.initInterceptor()
	svr.initRemoteServer()
	governor.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		if r.URL.Query().Get("pretty") == "true" {
			encoder.SetIndent("", "    ")
		}
		result := map[string]interface{}{
			"appName":  pkg.Name(),
			"services": svr.servicesDesc,
		}
		_ = encoder.Encode(result)
	})

	return svr
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
			logger.ErrorField("fault to stop server",
				logger.String("protocol", item.Info().Protocol), logger.Err(err))
		}
	}
	return multierr.Combine(errs...)
}

func (s *server) Serve(startFlag chan<- struct{}) error {
	s.mu.Lock()
	if s.state == serverStateClosing {
		s.mu.Unlock()
		return errors.New("server stopped")
	}
	if s.state == serverStateRunning {
		s.mu.Unlock()
		return errors.New("server already serve")
	}
	s.state = serverStateRunning
	s.mu.Unlock()
	for _, svr := range s.servers {
		if err := s.serve(svr); err != nil {
			return err
		}
	}
	startFlag <- struct{}{}
	s.serverWG.Wait()
	return nil
}

func (s *server) Endpoints() []Endpoint {
	endpoints := make([]Endpoint, len(s.servers))
	for i, item := range s.servers {
		e := item.Info()
		endpoints[i] = &serverInfo{
			scheme:   e.Protocol,
			address:  e.Address,
			metadata: e.Attr,
		}
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
			logger.FatalField("not found server builder",
				logger.String("protocol", protocol))
		}
		svr, err := builder(s.handleStream)
		if err != nil {
			logger.FatalField("fault to new remote server",
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
	s.registerServiceDesc(sd)
	s.registerServiceInfo(sd, ss)
}

func (s *server) registerServiceInfo(sd *ServiceDesc, ss interface{}) {
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

func (s *server) registerServiceDesc(desc *ServiceDesc) {
	methods := make([]methodInfo, 0, len(desc.Methods)+len(desc.Streams))
	for _, item := range desc.Methods {
		methods = append(methods, methodInfo{
			MethodName: item.MethodName,
		})
	}
	for _, item := range desc.Streams {
		methods = append(methods, methodInfo{
			MethodName:    item.StreamName,
			ServerStreams: item.ServerStreams,
			ClientStreams: item.ClientStreams,
		})
	}
	s.servicesDesc[desc.ServiceName] = methods
}

func (s *server) serve(svr remote.Server) error {
	err := svr.Start()
	if err != nil {
		logger.ErrorField("the server was ended forcefully",
			logger.String("protocol", svr.Info().Protocol), logger.Err(err))
		return err
	}
	logger.InfoField("server started", logger.String("protocol", svr.Info().Protocol), logger.String("endpoint", svr.Info().Address))
	s.serverWG.Add(1)
	go func() {
		defer s.serverWG.Done()
		if err = svr.Handle(); err != nil {
			logger.ErrorField("fault to handle channel",
				logger.String("protocol", svr.Info().Protocol), logger.Err(err))
		}
	}()
	return nil
}

func (s *server) handleStream(ss remote.ServerStream) {
	sm := ss.Method()
	if sm != "" && sm[0] == '/' {
		sm = sm[1:]
	}
	pos := strings.LastIndex(sm, "/")
	if pos == -1 {
		ss.Finish(nil, status.Errorf(code.Code_UNIMPLEMENTED, fmt.Sprintf("malformed method name: %q", sm)))
		return
	}
	service := sm[:pos]
	method := sm[pos+1:]

	srv, knownService := s.services[service]
	if knownService {
		if md, ok := srv.Methods[method]; ok {
			s.processUnaryRPC(md, srv, ss)
			return
		}
		if sd, ok := srv.Streams[method]; ok {
			s.processStreamRpc(sd, srv, ss)
			return
		}
	}
	var errDesc string
	if !knownService {
		errDesc = fmt.Sprintf("unknown service %v", service)
	} else {
		errDesc = fmt.Sprintf("unknown method %v for service %v", method, service)
	}
	ss.Finish(nil, status.Errorf(code.Code_UNIMPLEMENTED, errDesc))
	return
}

func (s *server) processUnaryRPC(desc *MethodDesc, srv *ServiceInfo, ss remote.ServerStream) {
	var (
		reply any
		err   error
	)
	defer func() {
		ss.Finish(reply, err)
	}()
	if err = ss.Start(false, false); err != nil {
		return
	}

	ctx := metadata.WithStreamContext(ss.Context())
	reply, err = desc.Handler(srv.ServiceImpl, ctx, ss.RecvMsg, s.unaryInterceptor)
	if header, ok := metadata.FromHeaderCtx(ctx); ok {
		_ = ss.SetHeader(header)
	}
	if trailer, ok := metadata.FromTrailerCtx(ctx); ok {
		ss.SetTrailer(trailer)
	}
	return
}

func (s *server) processStreamRpc(desc *stream.StreamDesc, srv *ServiceInfo, ss remote.ServerStream) {
	var err error
	defer func() {
		ss.Finish(nil, err)
	}()
	if err = ss.Start(desc.ClientStreams, desc.ServerStreams); err != nil {
		return
	}
	si := &interceptor.StreamServerInfo{
		FullMethod:     ss.Method(),
		IsClientStream: desc.ClientStreams,
		IsServerStream: desc.ServerStreams,
	}
	err = s.streamInterceptor(srv.ServiceImpl, ss, si, desc.Handler)
	return
}
