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

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/stats"

	"github.com/imkuqin-zw/yggdrasil/internal/backoff"
	"github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/imkuqin-zw/yggdrasil/pkg/stream"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xsync"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/genproto/googleapis/rpc/code"
)

var (
	ErrClientClosing = status.Errorf(code.Code_CANCELLED, "the client is closing")
)

type Client interface {
	// Invoke performs a unary RPC and returns after the response is received into reply.
	Invoke(ctx context.Context, method string, args, reply interface{}) error
	// NewStream begins a streaming RPC.
	NewStream(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error)
	// Close destroy the client resource.
	Close() error
}

type instance struct {
	Address  string
	Protocol string
	Metadata map[string]interface{}
}

func (i instance) GetAddress() string {
	return i.Address
}

func (i instance) GetProtocol() string {
	return i.Protocol
}

func (i instance) GetMetadata() map[string]interface{} {
	return i.Metadata
}

type pickSnap struct {
	version   int64
	balancer  balancer.Balancer
	remoteCli map[string]remote.Client
}

type clientStream struct {
	desc *stream.StreamDesc
	stream.ClientStream
	report func(err error)
}

func (c *clientStream) SendMsg(m interface{}) error {
	err := c.ClientStream.SendMsg(m)
	if err != nil && err != io.EOF {
		c.report(err)
	}
	return err
}

func (c *clientStream) RecvMsg(m interface{}) error {
	err := c.ClientStream.RecvMsg(m)
	if !c.desc.ServerStreams {
		if header, _ := c.Header(); header != nil {
			_ = metadata.SetHeader(c.Context(), header)
		}
		if trailer := c.Trailer(); trailer != nil {
			_ = metadata.SetTrailer(c.Context(), trailer)
		}
	}
	if err != nil && err != io.EOF && !c.desc.ServerStreams {
		c.report(err)
	}
	return err
}

type client struct {
	ctx               context.Context
	serviceName       string
	configChange      chan config.WatchEvent
	transportBackoff  backoff.Strategy
	snapVersion       atomic.Int64
	pickSnap          pickSnap
	resolvedEvent     *xsync.Event
	resolver          resolver.Resolver
	balancer          balancer.Balancer
	remoteCli         map[string]remote.Client
	unaryInterceptor  interceptor.UnaryClientInterceptor
	streamInterceptor interceptor.StreamClientInterceptor
	statsHandler      stats.Handler
}

func NewClient(ctx context.Context, serviceName string) (Client, error) {
	cli := &client{
		ctx:           ctx,
		serviceName:   serviceName,
		configChange:  make(chan config.WatchEvent, 1),
		remoteCli:     map[string]remote.Client{},
		resolvedEvent: xsync.NewEvent(),
		statsHandler:  stats.GetClientHandler(),
	}
	bc := backoff.DefaultConfig
	bc.BaseDelay = time.Millisecond * 50
	cli.transportBackoff = backoff.Exponential{Config: bc}
	cfgKey := fmt.Sprintf(config.KeyClientInstance, serviceName)
	cfg := config.ValueToValues(config.Get(cfgKey))
	if err := cli.initResolverAndBalancer(cfg); err != nil {
		return nil, err
	}
	cli.initInterceptor()
	xgo.Go(cli.watchConfigChange, nil)
	if err := config.AddWatcher(cfgKey, cli.notifyConfigChange); err != nil {
		return nil, err
	}
	return cli, nil
}

func (c *client) initResolverAndBalancer(cfg config.Values) error {
	balancerName := cfg.Get(config.KeySingleBalancer).String("round_robin")
	balancerBuilder, err := balancer.GetBuilder(balancerName)
	if err != nil {
		return err
	}
	c.balancer = balancerBuilder(c.serviceName)
	resolverName := cfg.Get(config.KeySingleResolver).String()
	if resolverName != "" {
		r, err := resolver.GetResolver(resolverName)
		if err != nil {
			return err
		}
		c.resolver = r
		if err := r.AddWatch(c.serviceName); err != nil {
			return err
		}
	} else {
		c.handlePickConfig(cfg)
		c.resolvedEvent.Fire()
	}
	return nil
}

func (c *client) initInterceptor() {
	unaryIntNames := xarray.RemoveEmptyStrings(xarray.RemoveReplaceStrings(append(
		strings.Split(config.Get(config.KeyIntUnaryClient).String(), ","),
		strings.Split(config.Get(fmt.Sprintf(config.KeyClientUnaryInt, c.serviceName)).String(), ",")...,
	)))
	c.unaryInterceptor = interceptor.ChainUnaryClientInterceptors(c.serviceName, unaryIntNames)
	streamIntNames := xarray.RemoveEmptyStrings(xarray.RemoveReplaceStrings(append(
		strings.Split(config.Get(config.KeyIntStreamClient).String(), ","),
		strings.Split(config.Get(fmt.Sprintf(config.KeyClientStreamInt, c.serviceName)).String(), ",")...,
	)))
	c.streamInterceptor = interceptor.ChainStreamClientInterceptors(c.serviceName, streamIntNames)
}

func (c *client) handleConfig(value config.Value) {
	cfg := config.ValueToValues(value)
	g := errgroup.Group{}
	g.Go(func() error {
		c.handlePickConfig(cfg)
		return nil
	})
	_ = g.Wait()
}

func (c *client) handlePickConfig(cfg config.Values) {
	endpoints := make([]instance, 0)
	if err := cfg.Get(config.KeySingleEndpoints).Scan(&endpoints); err != nil {
		logger.ErrorField("fault to load client config", logger.Err(err))
		return
	}
	remoteCli := make(map[string]remote.Client, len(endpoints))
	for _, item := range endpoints {
		if cli, ok := c.remoteCli[item.Address]; ok {
			remoteCli[item.Address] = cli
			continue
		}
		builder := remote.GetClientBuilder(item.Protocol)
		if builder == nil {
			logger.Warnf("not found client builder, protocol: %s", item.Protocol)
			continue
		}
		cli := builder(c.ctx, c.serviceName, item, c.statsHandler)
		if cli != nil {
			remoteCli[item.Address] = cli
		}

	}
	needDel := make([]remote.Client, 0)
	for key, item := range c.remoteCli {
		if _, ok := remoteCli[key]; ok {
			continue
		}
		needDel = append(needDel, item)
	}

	b := c.balancer
	balancerName := cfg.Get(config.KeySingleBalancer).String("round_robin")
	if balancerName != c.balancer.Name() {
		balancerBuilder, err := balancer.GetBuilder(balancerName)
		if err != nil {
			logger.Warn(err.Error())
		} else {
			b = balancerBuilder(c.serviceName)
		}
	}
	b.Update(cfg)
	c.remoteCli = remoteCli
	c.balancer = b
	version := c.snapVersion.Add(1)
	c.pickSnap = pickSnap{balancer: b, remoteCli: remoteCli, version: version}
	for _, item := range needDel {
		_ = item.Close()
	}
	if len(endpoints) != 0 {
		c.resolvedEvent.Fire()
	}
}

func (c *client) watchConfigChange() {
	var version uint64
	for {
		select {
		case <-c.ctx.Done():
			return
		case e, ok := <-c.configChange:
			if !ok {
				return
			}
			if e.Version() > version {
				c.handleConfig(e.Value())
				version = e.Version()
			}
		}
	}
}

func (c *client) notifyConfigChange(event config.WatchEvent) {
	for {
		select {
		case <-c.configChange:
		default:
		}
		break
	}
	c.configChange <- event
}

func (c *client) waitForResolved(ctx context.Context) error {
	if c.resolvedEvent.HasFired() {
		return nil
	}
	select {
	case <-c.resolvedEvent.Done():
		return nil
	case <-ctx.Done():
		return status.FromContextError(ctx.Err())
	case <-c.ctx.Done():
		return status.Errorf(code.Code_CANCELLED, "the client is closing")
	}
}

func (c *client) newConnStream(ctx context.Context, picker balancer.Picker, snap pickSnap, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
	r, err := picker.Next(balancer.RpcInfo{
		Ctx:    ctx,
		Method: method,
	})
	if err != nil {
		return nil, err
	}
	cli, ok := snap.remoteCli[r.Endpoint().GetAddress()]
	if !ok || cli == nil {
		return nil, status.Errorf(code.Code_UNAVAILABLE, "server cannot connect")
	}
	st, err := cli.NewStream(ctx, desc, method)
	if err != nil {
		r.Report(err)
		return nil, err
	}
	return &clientStream{
		desc:         desc,
		ClientStream: st,
		report:       r.Report,
	}, nil
}

func (c *client) newStream(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
	if err := c.waitForResolved(ctx); err != nil {
		return nil, err
	}
	retries := 0
	snap := c.pickSnap
	picker := snap.balancer.GetPicker()
	for {
		st, err := c.newConnStream(ctx, picker, snap, desc, method)
		if err == nil {
			return st, nil
		}
		logger.ErrorField("fault to new stream", logger.Err(err))
		if errors.Is(err, balancer.ErrNoAvailableInstance) {
			return nil, status.New(code.Code_UNAVAILABLE, err)
		}
		if retries > 3 {
			return nil, err
		}
		t := time.NewTimer(c.transportBackoff.Backoff(retries))
		select {
		case <-c.ctx.Done():
			return nil, ErrClientClosing
		case <-ctx.Done():
			return nil, err
		case <-t.C:
			retries++
		}
	}
}

func (c *client) invoke(ctx context.Context, method string, args, reply interface{}) error {
	cs, err := c.newStream(ctx, &stream.StreamDesc{ServerStreams: false, ClientStreams: false}, method)
	if err != nil {
		return err
	}
	if err = cs.SendMsg(args); err != nil {
		return err
	}
	err = cs.RecvMsg(reply)
	return err
}

func (c *client) Invoke(ctx context.Context, method string, args, reply interface{}) error {
	ctx = metadata.WithStreamContext(ctx)
	return c.unaryInterceptor(ctx, method, args, reply, c.invoke)
}

func (c *client) NewStream(ctx context.Context, desc *stream.StreamDesc, method string) (stream.ClientStream, error) {
	return c.streamInterceptor(ctx, desc, method, c.newStream)
}

func (c *client) Close() error {
	var mErr []error
	if err := config.DelWatcher(fmt.Sprintf(config.KeyClientInstance, c.serviceName), c.notifyConfigChange); err != nil {
		mErr = append(mErr, err)
	}
	if c.resolver != nil {
		if err := c.resolver.DelWatch(c.serviceName); err != nil {
			mErr = append(mErr, err)
		}
	}
	if len(mErr) > 0 {
		return multierr.Combine(mErr...)
	}
	return nil
}
