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

package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/encoding"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/encoding/proto"
	transport2 "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/transport"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/transport/keepalive"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/credentials"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"google.golang.org/genproto/googleapis/rpc/code"
)

func init() {
	remote.RegisterServerBuilder("grpc", newServer)
}

const (
	defaultServerMaxReceiveMessageSize = 1024 * 1024 * 4
	defaultServerMaxSendMessageSize    = math.MaxInt32
)

type serverOptions struct {
	Network               string
	Address               string
	CredsProto            string
	CodeProto             string
	MaxConcurrentStreams  uint32
	MaxReceiveMessageSize int
	MaxSendMessageSize    int
	KeepaliveParams       keepalive.ServerParameters
	KeepalivePolicy       keepalive.EnforcementPolicy
	InitialWindowSize     int32
	InitialConnWindowSize int32
	WriteBufferSize       int
	ReadBufferSize        int
	ConnectionTimeout     time.Duration
	MaxHeaderListSize     *uint32
	HeaderTableSize       *uint32

	Attr map[string]string

	creds credentials.TransportCredentials
	codec encoding.Codec
}

func (opts *serverOptions) SetDefault() error {
	var err error
	if opts.Network == "" {
		opts.Network = "tcp"
	}
	if opts.Address == "" {
		opts.Address, err = xnet.Extract(opts.Address)
		if err != nil {
			return err
		}
		opts.Address = fmt.Sprintf("%s:0", opts.Address)
	}
	if opts.CodeProto == "" {
		opts.CodeProto = proto.Name
	}
	if opts.MaxReceiveMessageSize == 0 {
		opts.MaxReceiveMessageSize = defaultServerMaxReceiveMessageSize
	}
	if opts.MaxSendMessageSize == 0 {
		opts.MaxSendMessageSize = defaultServerMaxSendMessageSize
	}
	if opts.WriteBufferSize == 0 {
		opts.WriteBufferSize = defaultWriteBufSize
	}
	if opts.ReadBufferSize == 0 {
		opts.ReadBufferSize = defaultReadBufSize
	}
	if opts.ConnectionTimeout == 0 {
		opts.ConnectionTimeout = 120 * time.Second
	}
	return err
}

type server struct {
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	address   string
	lis       net.Listener
	serve     bool
	stopped   bool
	stoppedCh chan struct{}
	drain     bool
	cv        *sync.Cond // signaled when connections close for GracefulStop
	// conns contains all active server transports. It is a map keyed on a
	// listener address with the value being the set of active transports
	// belonging to that listener.
	conns   map[string]map[transport2.ServerTransport]bool
	opts    serverOptions
	handle  remote.MethodHandle
	serveWG sync.WaitGroup
}

func newServer(handle remote.MethodHandle) (remote.Server, error) {
	opts := serverOptions{}
	if err := config.Get(fmt.Sprintf(config.KeyRemoteProto, scheme)).Scan(&opts); err != nil {
		return nil, err
	}
	if err := opts.SetDefault(); err != nil {
		return nil, err
	}
	opts.codec = encoding.GetCodec(opts.CodeProto)
	s := &server{
		stoppedCh: make(chan struct{}),
		conns:     make(map[string]map[transport2.ServerTransport]bool),
		opts:      opts,
		handle:    handle,
	}
	s.cv = sync.NewCond(&s.mu)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s, nil
}

func (s *server) serveStreams(st transport2.ServerTransport) {
	defer st.Close()
	var wg sync.WaitGroup
	st.HandleStreams(func(stream *transport2.Stream) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleStream(st, stream)
		}()
	}, func(ctx context.Context, s string) context.Context {
		return ctx
	})
	wg.Wait()
}

func (s *server) handleStream(t transport2.ServerTransport, stream *transport2.Stream) {
	ctx := stream.Context()
	ss := &serverStream{
		ctx:                   ctx,
		t:                     t,
		s:                     stream,
		p:                     &parser{r: stream},
		codec:                 s.getCodec(stream.ContentSubtype()),
		maxReceiveMessageSize: s.opts.MaxReceiveMessageSize,
		maxSendMessageSize:    s.opts.MaxSendMessageSize,
	}

	// If dc is set and matches the stream's compression, use it.  Otherwise, try
	// to find a matching registered compressor for decomp.
	rc := stream.RecvCompress()
	if rc != "" && rc != encoding.Identity {
		ss.decomp = encoding.GetCompressor(rc)
		if ss.decomp == nil {
			st := status.Errorf(code.Code_UNIMPLEMENTED, fmt.Sprintf("grpc: Decompressor is not installed for grpc-encoding %q", rc))
			t.WriteStatus(ss.s, st)
		}
	}

	// If cp is set, use it.  Otherwise, attempt to compress the response using
	// the incoming message compression method.
	//
	// NOTE: this needs to be ahead of all handling, https://github.com/grpc/grpc-go/issues/686.
	if rc := stream.RecvCompress(); rc != "" && rc != encoding.Identity {
		// Legacy compressor not specified; attempt to respond with same encoding.
		ss.comp = encoding.GetCompressor(rc)
		if ss.comp != nil {
			stream.SetSendCompress(rc)
		}
	}

	reply, isStream, err := s.handle(ctx, stream.Method(), ss)
	if err != nil {
		t.WriteStatus(stream, status.FromError(err))
		return
	}
	if !isStream {
		opts := &transport2.Options{Last: true}

		if err := s.sendResponse(t, stream, reply, opts, ss.comp); err != nil {
			if err == io.EOF {
				// The entire stream is done (for unary RPC only).
				return
			}
			if sts, ok := status.CoverError(err); ok {
				if e := t.WriteStatus(stream, sts); e != nil {
				}
			} else {
				switch st := err.(type) {
				case transport2.ConnectionError:
					// Nothing to do here.
				default:
					panic(fmt.Sprintf("grpc: Unexpected reason (%T) from sendResponse: %v", st, st))
				}
			}
			return
		}
		// TODO: Should we be logging if writing status failed here, like above?
		// Should the logging be in WriteStatus?  Should we ignore the WriteStatus
		// reason or allow the stats handler to see it?
		err = t.WriteStatus(stream, status.New(code.Code_OK, nil))
	}
}

func (s *server) getCodec(contentSubtype string) encoding.Codec {
	if s.opts.codec != nil {
		return s.opts.codec
	}
	if contentSubtype == "" {
		return encoding.GetCodec(proto.Name)
	}
	codec := encoding.GetCodec(contentSubtype)
	if codec == nil {
		return encoding.GetCodec(proto.Name)
	}
	return codec
}

func (s *server) sendResponse(t transport2.ServerTransport, stream *transport2.Stream, msg interface{}, opts *transport2.Options, comp encoding.Compressor) error {
	data, err := encode(s.getCodec(stream.ContentSubtype()), msg)
	if err != nil {
		return err
	}
	compData, err := compress(data, comp)
	if err != nil {
		return err
	}
	hdr, payload := msgHeader(data, compData)
	// TODO(dfawley): should we be checking len(data) instead?
	if len(payload) > s.opts.MaxSendMessageSize {
		return status.Errorf(code.Code_RESOURCE_EXHAUSTED, fmt.Sprintf("grpc: trying to send message larger than max (%d vs. %d)", len(payload), s.opts.MaxSendMessageSize))
	}
	err = t.Write(stream, hdr, payload, opts)
	return err
}

func (s *server) newTransport(c net.Conn) transport2.ServerTransport {
	config := &transport2.ServerConfig{
		MaxStreams:            s.opts.MaxConcurrentStreams,
		ConnectionTimeout:     s.opts.ConnectionTimeout,
		Credentials:           s.opts.creds,
		KeepaliveParams:       s.opts.KeepaliveParams,
		KeepalivePolicy:       s.opts.KeepalivePolicy,
		InitialWindowSize:     s.opts.InitialWindowSize,
		InitialConnWindowSize: s.opts.InitialConnWindowSize,
		WriteBufferSize:       s.opts.WriteBufferSize,
		ReadBufferSize:        s.opts.ReadBufferSize,
		MaxHeaderListSize:     s.opts.MaxHeaderListSize,
		HeaderTableSize:       s.opts.HeaderTableSize,
	}
	st, err := transport2.NewServerTransport(c, config)
	if err != nil {
		// ErrConnDispatched means that the connection was dispatched away from
		// gRPC; those connections should be left open.
		if err != credentials.ErrConnDispatched {
			// Don't log on ErrConnDispatched and io.EOF to prevent log spam.
			c.Close()
		}
		return nil
	}
	return st
}

func (s *server) Serve() (<-chan error, error) {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil, errors.New("server had already stopped")
	}
	if s.serve {
		return nil, errors.New("server had already serve")
	}
	s.serve = true
	s.mu.Unlock()
	lis, err := net.Listen(s.opts.Network, s.opts.Address)
	if err != nil {
		return nil, err
	}
	s.address = lis.Addr().String()
	s.lis = lis
	var tempDelay time.Duration
	ch := make(chan error)
	go func() {
		var err error
		defer func() {
			<-s.stoppedCh
			if err != nil {
				ch <- err
			}
			close(ch)
		}()
		for {
			rawConn, acceptErr := s.lis.Accept()
			if acceptErr != nil {
				if ne, ok := acceptErr.(interface{ Temporary() bool }); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					remote.Logger.Warnf("Accept reason: %v; retrying in %v", acceptErr, tempDelay)
					timer := time.NewTimer(tempDelay)
					select {
					case <-timer.C:
					case <-s.ctx.Done():
						timer.Stop()
						return
					}
					continue
				}
				s.mu.Lock()
				if s.stopped {
					s.mu.Unlock()
					return
				}
				s.mu.Unlock()
				remote.Logger.Errorf("done serving; Accept = %v", acceptErr)
				err = acceptErr
				return
			}
			tempDelay = 0
			// Start a new goroutine to deal with rawConn so we don't stall this Accept
			// loop goroutine.
			//
			// Make sure we account for the goroutine so GracefulStop doesn't nil out
			// s.conns before this conn can be added.
			s.serveWG.Add(1)
			//TODO: add goroutine pool
			go func() {
				s.handleRawConn(lis.Addr().String(), rawConn)
				s.serveWG.Done()
			}()
		}
	}()
	return ch, nil
}

func (s *server) Stop() error {
	s.mu.Lock()
	if !s.serve {
		s.stopped = true
		s.mu.Unlock()
		return nil
	}
	if s.stopped {
		s.mu.Unlock()
		<-s.stoppedCh
		return nil
	}
	s.stopped = true
	s.mu.Unlock()
	s.cancel()
	if s.lis != nil {
		_ = s.lis.Close()
	}
	if !s.drain {
		for _, conns := range s.conns {
			for st := range conns {
				st.Drain()
			}
		}
		s.drain = true
	}

	// Wait for serving threads to be ready to exit.  Only then can we be sure no
	// new conns will be created.
	s.serveWG.Wait()
	s.mu.Lock()
	conns := s.conns
	s.conns = nil
	s.mu.Unlock()
	for len(conns) != 0 {
		s.cv.Wait()
	}
	close(s.stoppedCh)
	return nil
}

func (s *server) Info() remote.ServerInfo {
	return remote.ServerInfo{
		Address:  s.address,
		Protocol: scheme,
		Attr:     s.opts.Attr,
	}
}

// handleRawConn forks a goroutine to handle a just-accepted connection that
// has not had any I/O performed on it yet.
func (s *server) handleRawConn(lisAddr string, rawConn net.Conn) {
	if s.stopped {
		rawConn.Close()
		return
	}
	rawConn.SetDeadline(time.Now().Add(s.opts.ConnectionTimeout))

	// Finish handshaking (HTTP2)
	st := s.newTransport(rawConn)
	rawConn.SetDeadline(time.Time{})
	if st == nil {
		return
	}

	if !s.addConn(lisAddr, st) {
		return
	}
	go func() {
		s.serveStreams(st)
		s.removeConn(lisAddr, st)
	}()
}

func (s *server) addConn(addr string, st transport2.ServerTransport) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conns == nil {
		st.Close()
		return false
	}
	if s.drain {
		// Transport added after we drained our existing conns: drain it
		// immediately.
		st.Drain()
	}

	if s.conns[addr] == nil {
		// Create a map entry if this is the first connection on this listener.
		s.conns[addr] = make(map[transport2.ServerTransport]bool)
	}
	s.conns[addr][st] = true
	return true
}

func (s *server) removeConn(addr string, st transport2.ServerTransport) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns := s.conns[addr]
	if conns != nil {
		delete(conns, st)
		if len(conns) == 0 {
			// If the last connection for this address is being removed, also
			// remove the map entry corresponding to the address. This is used
			// in GracefulStop() when waiting for all connections to be closed.
			delete(s.conns, addr)
		}
		s.cv.Broadcast()
	}
}

// serverStream implements a server side Stream.
type serverStream struct {
	ctx   context.Context
	t     transport2.ServerTransport
	s     *transport2.Stream
	p     *parser
	codec encoding.Codec

	comp   encoding.Compressor
	decomp encoding.Compressor

	maxReceiveMessageSize int
	maxSendMessageSize    int

	mu sync.Mutex // protects trInfo.tr after the service handler runs.
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

func (ss *serverStream) SetHeader(md metadata.MD) error {
	if md.Len() == 0 {
		return nil
	}
	return ss.s.SetHeader(md)
}

func (ss *serverStream) SendHeader(md metadata.MD) error {
	err := ss.t.WriteHeader(ss.s, md)
	return err
}

func (ss *serverStream) SetTrailer(md metadata.MD) {
	if md.Len() == 0 {
		return
	}
	ss.s.SetTrailer(md)
}

func (ss *serverStream) SendMsg(m interface{}) (err error) {
	defer func() {
		if err != nil && err != io.EOF {
			st := status.FromError(toRPCErr(err))
			ss.t.WriteStatus(ss.s, st)
			// Non-user specified status was sent out. This should be an reason
			// case (as a server side Cancel maybe).
			//
			// This is not handled specifically now. User will return a final
			// status from the service handler, we will log that reason instead.
			// This behavior is similar to an interceptor.
		}
	}()

	// load hdr, payload, data
	hdr, payload, _, err := prepareMsg(m, ss.codec, ss.comp)
	if err != nil {
		return err
	}

	// TODO(dfawley): should we be checking len(data) instead?
	if len(payload) > ss.maxSendMessageSize {
		return status.Errorf(code.Code_RESOURCE_EXHAUSTED, fmt.Sprintf("trying to send message larger than max (%d vs. %d)", len(payload), ss.maxSendMessageSize))
	}
	if err := ss.t.Write(ss.s, hdr, payload, &transport2.Options{Last: false}); err != nil {
		return toRPCErr(err)
	}
	return nil
}

func (ss *serverStream) RecvMsg(m interface{}) (err error) {
	defer func() {
		if err != nil && err != io.EOF {
			st := status.FromError(toRPCErr(err))
			ss.t.WriteStatus(ss.s, st)
			// Non-user specified status was sent out. This should be an reason
			// case (as a server side Cancel maybe).
			//
			// This is not handled specifically now. User will return a final
			// status from the service handler, we will log that reason instead.
			// This behavior is similar to an interceptor.
		}
	}()
	var payInfo *payloadInfo
	if err := recv(ss.p, ss.codec, ss.s, m, ss.maxReceiveMessageSize, payInfo, ss.decomp); err != nil {
		if err == io.EOF {
			return err
		}
		if err == io.ErrUnexpectedEOF {
			err = status.Errorf(code.Code_INTERNAL, io.ErrUnexpectedEOF.Error())
		}
		return toRPCErr(err)
	}
	return nil
}
