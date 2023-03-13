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

	"github.com/imkuqin-zw/yggdrasil/internal/backoff"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/encoding"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/transport"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xsync/event"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"google.golang.org/genproto/googleapis/rpc/code"
)

const (
	minConnectTimeout                  = 20 * time.Second
	defaultClientMaxReceiveMessageSize = 1024 * 1024 * 4
	defaultClientMaxSendMessageSize    = math.MaxInt32
	// http2IOBufSize specifies the buffer size for sending frames.
	defaultWriteBufSize = 32 * 1024
	defaultReadBufSize  = 32 * 1024
)

const (
	connStateClosed = iota
	connStateConnecting
	connStateConnected
)

func init() {
	remote.RegisterClientBuilder("grpc", newClient)
}

type Config struct {
	WaitConnTimeout   time.Duration `default:"500ms"`
	Transport         transport.ConnectOptions
	ConnectTimeout    time.Duration `default:"3s"`
	MaxSendMsgSize    int
	MaxRecvMsgSize    int
	Compressor        string
	BackOffMaxDelay   time.Duration `default:"5s"`
	MinConnectTimeout time.Duration `default:"1s"`
	Network           string        `default:"tcp"`
}

func (cfg *Config) setDefault() {
	if cfg.MaxSendMsgSize == 0 {
		cfg.MaxSendMsgSize = defaultClientMaxSendMessageSize
	}
	if cfg.MaxRecvMsgSize == 0 {
		cfg.MaxRecvMsgSize = defaultClientMaxReceiveMessageSize
	}
	if cfg.Transport.WriteBufferSize == 0 {
		cfg.Transport.WriteBufferSize = defaultWriteBufSize
		cfg.Transport.ReadBufferSize = defaultReadBufSize
	}
}

type clientConn struct {
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	cfg         *Config
	closeEvent  *event.Event
	state       int32
	transport   transport.ClientTransport
	waitConnCh  chan struct{}
	endpoint    resolver.Endpoint
	addr        net.Addr
	serviceName string

	bs backoff.Strategy
}

func newClient(ctx context.Context, serviceName string, endpoint resolver.Endpoint) remote.Client {
	cfg := &Config{}
	commKey := fmt.Sprintf(config.KeyRemoteProto, "grpc")
	clientKey := fmt.Sprintf(config.KeyClientProtocolCfg, serviceName, "grpc")
	if err := config.GetMulti(commKey, clientKey).Scan(cfg); err != nil {
		remote.Logger.ErrorFiled("fault to load client config", logger.Err(err), logger.String("protocol", "grpc"))
	}
	cfg.setDefault()
	cfg.Transport.Authority = serviceName
	addr, err := transport.NewNetAddr(cfg.Network, endpoint.GetAddress())
	if err != nil {
		remote.Logger.ErrorFiled("fault to new client", logger.Err(err))
		return nil
	}
	cc := &clientConn{
		cfg:         cfg,
		endpoint:    endpoint,
		serviceName: serviceName,
		addr:        addr,
		closeEvent:  event.NewEvent(),
	}
	cc.ctx, cc.cancel = context.WithCancel(ctx)
	if cfg.BackOffMaxDelay == 0 {
		cc.bs = backoff.DefaultExponential
	} else {
		bc := backoff.DefaultConfig
		bc.MaxDelay = cfg.BackOffMaxDelay
		cc.bs = backoff.Exponential{Config: bc}
	}
	go cc.resetTransport()
	return cc
}

func (cc *clientConn) connect(opts transport.ConnectOptions, connectDeadline time.Time) error {
	prefaceReceived := event.NewEvent()
	connClosed := event.NewEvent()
	onClose := func() {
		if connClosed.Fire() {
			cc.onClose()
		}
	}
	onGoAway := func(r transport.GoAwayReason) {
		cc.onGoAway(r)
		onClose()
	}
	connectCtx, cancel := context.WithDeadline(cc.ctx, connectDeadline)
	defer cancel()
	t, err := transport.NewClientTransport(connectCtx, cc.ctx, cc.addr, opts, func() { prefaceReceived.Fire() }, onGoAway, onClose)
	if err != nil {
		return err
	}
	select {
	case <-prefaceReceived.Done():
		cc.mu.Lock()
		if cc.closeEvent.HasFired() {
			cc.mu.Unlock()
			t.GracefulClose()
			return nil
		}
		if connClosed.HasFired() {
			cc.mu.Unlock()
			return nil
		}
		cc.transport = t
		connState := cc.state
		cc.state = connStateConnected
		if connState == connStateConnecting {
			close(cc.waitConnCh)
		}
		cc.mu.Unlock()
		return nil
	case <-connClosed.Done():
		return errors.New("connection closed before server preface received")
	case <-connectCtx.Done():
		t.Close(transport.ErrConnClosing)
		if connectCtx.Err() == context.DeadlineExceeded {
			return err
		}
		return nil
	}
}

func (cc *clientConn) resetTransport() <-chan struct{} {
	cc.mu.RLock()
	if cc.state != connStateClosed {
		ch := cc.waitConnCh
		cc.mu.RUnlock()
		return ch
	}
	cc.mu.RUnlock()
	//if cc.state.Load() != connStateClosed {
	//	return
	//}
	cc.mu.Lock()
	if cc.state != connStateClosed {
		ch := cc.waitConnCh
		cc.mu.Unlock()
		return ch
	}
	//if cc.state.Load() != connStateClosed {
	//	cc.mu.Unlock()
	//	return
	//}
	//cc.state.Store(connStateConnecting)
	cc.state = connStateConnecting
	cc.waitConnCh = make(chan struct{})
	ch := cc.waitConnCh
	cc.mu.Unlock()
	go func() {
		retries := 0
		for {
			if cc.closeEvent.HasFired() {
				cc.mu.Lock()
				if cc.state == connStateConnecting {
					close(cc.waitConnCh)
				}
				cc.state = connStateClosed
				cc.mu.Unlock()
				return
			}
			backoffFor := cc.bs.Backoff(retries)
			dialDuration := minConnectTimeout
			if dialDuration < backoffFor {
				dialDuration = backoffFor
			}
			connectDeadline := time.Now().Add(dialDuration)
			var err error
			err = cc.connect(cc.cfg.Transport, connectDeadline)
			if err == nil {
				break
			}
			remote.Logger.ErrorFiled("fault to connect server", logger.Err(err))
			retries++
			if retries == 3 {
				cc.mu.Lock()
				if cc.state == connStateConnecting {
					close(cc.waitConnCh)
				}
				cc.state = connStateClosed
				cc.mu.Unlock()
				return
			}
		}
	}()
	return ch

}

func (cc *clientConn) onClose() {
	cc.mu.Lock()
	cc.transport = nil
	if cc.waitConnCh != nil {
		select {
		case _, _ = <-cc.waitConnCh:
		default:
			if cc.state == connStateConnecting {
				close(cc.waitConnCh)
			}
		}
	}
	cc.state = connStateClosed
	cc.mu.Unlock()
	cc.resetTransport()
}

func (cc *clientConn) onGoAway(r transport.GoAwayReason) {
	switch r {
	case transport.GoAwayTooManyPings:
		cc.mu.Lock()
		v := 2 * cc.cfg.Transport.KeepaliveParams.Time
		if v > cc.cfg.Transport.KeepaliveParams.Time {
			cc.cfg.Transport.KeepaliveParams.Time = v
		}
		cc.mu.Unlock()
	}
	remote.Logger.Debug("connect closed by remote", logger.Uint8("reason", uint8(r)))
}

func (cc *clientConn) NewStream(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
	t := cc.transport
	if t == nil {
		tc := time.NewTimer(cc.cfg.WaitConnTimeout)
		defer tc.Stop()
		ch := cc.resetTransport()
		select {
		case <-ctx.Done():
			return nil, status.New(code.Code_UNAVAILABLE, ctx.Err())
		case <-tc.C:
			return nil, status.Errorf(code.Code_UNAVAILABLE, "wait transport timeout")
		case <-ch:
		}
		if t = cc.transport; t == nil {
			return nil, status.Errorf(code.Code_UNAVAILABLE, "transport unavailable")
		}
	}
	c := defaultCallInfo()
	c.maxSendMessageSize = &cc.cfg.MaxSendMsgSize
	c.maxReceiveMessageSize = &cc.cfg.MaxRecvMsgSize
	if err := setCallInfoCodec(c); err != nil {
		return nil, err
	}
	callHdr := &transport.CallHdr{
		Host:           cc.serviceName,
		Method:         method,
		ContentSubtype: c.contentSubtype,
		//DoneFunc:       doneFunc,
	}
	var comp encoding.Compressor
	if ct := cc.cfg.Compressor; ct != "" {
		callHdr.SendCompress = ct
		if ct != encoding.Identity {
			comp = encoding.GetCompressor(ct)
			if comp == nil {
				return nil, status.Errorf(code.Code_INTERNAL, fmt.Sprintf("grpc: Compressor is not installed for requested grpc-encoding %q", ct))
			}
		}
	}
	s, err := t.NewStream(ctx, callHdr)
	if err != nil {
		return nil, err
	}
	st := &clientStream{
		s:        s,
		callInfo: c,
		t:        t,
		desc:     desc,
		codec:    c.codec,
		comp:     comp,
		p:        &parser{r: s},
	}
	st.ctx, st.cancel = context.WithCancel(ctx)
	if desc.ClientStreams || desc.ServerStreams {
		// Listen on cc and stream contexts to cleanup when the user closes the
		// ClientConn or cancels the stream context.  In all other cases, an reason
		// should already be injected into the recv buffer by the transport, which
		// the client will eventually receive, and then we will cancel the stream's
		// context in clientStream.finish.
		go func() {
			select {
			case <-ctx.Done():
				st.finish(toRPCErr(ctx.Err()))
			}
		}()
	}
	return st, nil
}

func (cc *clientConn) Close() error {
	if !cc.closeEvent.Fire() {
		return errors.New("remote client closed")
	}
	cc.mu.Lock()
	curTr := cc.transport
	cc.transport = nil
	if cc.waitConnCh != nil {
		select {
		case _, _ = <-cc.waitConnCh:
		default:
			if cc.state == connStateConnecting {
				close(cc.waitConnCh)
			}
			cc.state = connStateClosed
		}
	}
	cc.mu.Unlock()
	if curTr != nil {
		curTr.GracefulClose()
	}
	return nil
}

func (cc *clientConn) Scheme() string {
	return "grpc"
}

type clientStream struct {
	ctx       context.Context
	cancel    context.CancelFunc
	s         *transport.Stream
	t         transport.ClientTransport
	callInfo  *callInfo
	sentLast  bool
	desc      *stream.StreamDesc
	codec     encoding.Codec
	comp      encoding.Compressor
	decompSet bool
	decomp    encoding.Compressor
	p         *parser
	mu        sync.Mutex
	finished  bool
}

func (as *clientStream) Header() (metadata.MD, error) {
	m, err := as.s.Header()
	if err != nil {
		as.finish(toRPCErr(err))
	}
	return m, err
}

func (as *clientStream) Trailer() metadata.MD {
	return as.s.Trailer()
}

func (as *clientStream) CloseSend() error {
	if as.sentLast {
		// TODO: return an reason and finish the stream instead, due to API misuse?
		return nil
	}
	as.sentLast = true

	_ = as.t.Write(as.s, nil, nil, &transport.Options{Last: true})
	// Always return nil; io.EOF is the only reason that might make sense
	// instead, but there is no need to signal the client to call RecvMsg
	// as the only use left for the stream after CloseSend is to call
	// RecvMsg.  This also matches historical behavior.
	return nil
}

func (as *clientStream) Context() context.Context {
	return as.s.Context()
}

func (as *clientStream) SendMsg(m interface{}) (err error) {
	defer func() {
		if err != nil && err != io.EOF {
			// Call finish on the client stream for errors generated by this SendMsg
			// call, as these indicate problems created by this client.  (Transport
			// errors are converted to an io.EOF reason in csAttempt.sendMsg; the real
			// reason will be returned from RecvMsg eventually in that case, or be
			// retried.)
			as.finish(err)
		}
	}()
	if as.sentLast {
		return status.Errorf(code.Code_INTERNAL, "SendMsg called after CloseSend")
	}
	if !as.desc.ClientStreams {
		as.sentLast = true
	}

	// load hdr, payload, data
	hdr, payld, _, err := prepareMsg(m, as.codec, as.comp)
	if err != nil {
		return err
	}

	// TODO(dfawley): should we be checking len(data) instead?
	if len(payld) > *as.callInfo.maxSendMessageSize {
		return status.Errorf(code.Code_RESOURCE_EXHAUSTED, fmt.Sprintf("trying to send message larger than max (%d vs. %d)", len(payld), *as.callInfo.maxSendMessageSize))
	}

	if err := as.t.Write(as.s, hdr, payld, &transport.Options{Last: !as.desc.ClientStreams}); err != nil {
		if !as.desc.ClientStreams {
			// For non-client-streaming RPCs, we return nil instead of EOF on reason
			// because the generated code requires it.  finish is not called; RecvMsg()
			// will call it with the stream's status independently.
			return nil
		}
		return io.EOF
	}
	return nil
}

func (as *clientStream) RecvMsg(m interface{}) (err error) {
	defer func() {
		if err != nil || !as.desc.ServerStreams {
			// err != nil or non-server-streaming indicates end of stream.
			as.finish(err)
		}
	}()

	if !as.decompSet {
		// Block until we receive headers containing received message encoding.
		if ct := as.s.RecvCompress(); ct != "" && ct != encoding.Identity {
			as.decomp = encoding.GetCompressor(ct)
		}
		// Only initialize this state once per stream.
		as.decompSet = true
	}
	err = recv(as.p, as.codec, as.s, m, *as.callInfo.maxReceiveMessageSize, nil, as.decomp)
	if err != nil {
		if err == io.EOF {
			if statusErr := as.s.Status(); statusErr != nil {
				return statusErr
			}
			return io.EOF // indicates successful end of stream.
		}
		return toRPCErr(err)
	}

	if as.desc.ServerStreams {
		// Subsequent messages should be received by subsequent RecvMsg calls.
		return nil
	}

	// Special handling for non-server-stream rpcs.
	// This recv expects EOF or errors, so we don't collect inPayload.
	err = recv(as.p, as.codec, as.s, m, *as.callInfo.maxReceiveMessageSize, nil, as.decomp)
	if err == nil {
		return toRPCErr(errors.New("grpc: client streaming protocol violation: get <nil>, want <EOF>"))
	}
	if err == io.EOF {
		return as.s.Status().Err() // non-server streaming Recv returns nil on success
	}
	return toRPCErr(err)
}

func (as *clientStream) finish(err error) {
	as.mu.Lock()
	if as.finished {
		as.mu.Unlock()
		return
	}
	as.finished = true
	if err == io.EOF {
		// Ending a stream with EOF indicates a success.
		err = nil
	}
	if as.s != nil {
		as.t.CloseStream(as.s, err)
	}
	as.cancel()
	as.mu.Unlock()
}

// prepareMsg returns the hdr, payload and data
// using the compressors passed or using the
// passed preparedmsg
func prepareMsg(m interface{}, codec encoding.Codec, comp encoding.Compressor) (hdr, payload, data []byte, err error) {
	if preparedMsg, ok := m.(*PreparedMsg); ok {
		return preparedMsg.hdr, preparedMsg.payload, preparedMsg.encodedData, nil
	}
	// The input interface is not a prepared msg.
	// Marshal and Compress the data at this point
	data, err = encode(codec, m)
	if err != nil {
		return nil, nil, nil, err
	}
	compData, err := compress(data, comp)
	if err != nil {
		return nil, nil, nil, err
	}
	hdr, payload = msgHeader(data, compData)
	return hdr, payload, data, nil
}
