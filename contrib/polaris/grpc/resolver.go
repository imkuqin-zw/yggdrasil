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

/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/imkuqin-zw/yggdrasil/contrib/polaris"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/api"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"google.golang.org/protobuf/proto"
)

type resolverBuilder struct {
}

// Scheme polaris scheme
func (rb *resolverBuilder) Scheme() string {
	return scheme
}

// Build Implement the Build method in the Resolver Builder interface,
// build a new Resolver resolution service address for the specified Target,
// and pass the polaris information to the balancer through attr
func (rb *resolverBuilder) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
) (resolver.Resolver, error) {
	dialOpts := &dialOpts{}
	key := fmt.Sprintf("yggdrasil.polaris.resolver.%s", target.URL.Host)
	if err := config.Scan(key, dialOpts); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	d := &polarisNamingResolver{
		ctx:     ctx,
		cancel:  cancel,
		cc:      cc,
		rn:      make(chan struct{}, 1),
		target:  target,
		options: dialOpts,
	}
	// d.wg.Add(1)
	go d.watcher()
	d.ResolveNow(resolver.ResolveNowOptions{})
	return d, nil
}

type polarisNamingResolver struct {
	ctx    context.Context
	cancel context.CancelFunc
	cc     resolver.ClientConn
	// rn channel is used by ResolveNow() to force an immediate resolution of the target.
	rn chan struct{}
	// wg          sync.WaitGroup
	options     *dialOpts
	target      resolver.Target
	balanceOnce sync.Once
}

// ResolveNow The method is called by the gRPC framework to resolve the target name
func (pr *polarisNamingResolver) ResolveNow(opt resolver.ResolveNowOptions) { // ??????resolve???????????????????????????
	select {
	case pr.rn <- struct{}{}:
	default:
	}
}

func getNamespace(options *dialOpts) string {
	namespace := polaris.DefaultNamespace
	if len(options.Namespace) > 0 {
		namespace = options.Namespace
	}
	return namespace
}

func (pr *polarisNamingResolver) lookup() (*resolver.State, api.ConsumerAPI, error) {
	sdkCtx, err := polaris.Context()
	if nil != err {
		return nil, nil, err
	}
	consumerAPI := api.NewConsumerAPIByContext(sdkCtx)
	instancesRequest := &api.GetInstancesRequest{}
	instancesRequest.Namespace = getNamespace(pr.options)
	instancesRequest.Service = pr.target.URL.Host
	if len(pr.options.DstMetadata) > 0 {
		instancesRequest.Metadata = pr.options.DstMetadata
	}
	sourceService := buildSourceInfo(pr.options)
	if sourceService != nil {
		// ?????????Conf????????????SourceService????????????????????????
		instancesRequest.SourceService = sourceService
	}
	instancesRequest.SkipRouteFilter = true
	resp, err := consumerAPI.GetInstances(instancesRequest)
	if nil != err {
		return nil, consumerAPI, err
	}
	state := &resolver.State{}
	for _, instance := range resp.Instances {
		if instance.GetProtocol() == "grpc" {
			state.Addresses = append(state.Addresses, resolver.Address{
				Addr:       fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
				Attributes: attributes.New(keyDialOptions, pr.options),
			})
		}
	}
	return state, consumerAPI, nil
}

func (pr *polarisNamingResolver) doWatch(
	consumerAPI api.ConsumerAPI) (model.ServiceKey, <-chan model.SubScribeEvent, error) {
	watchRequest := &api.WatchServiceRequest{}
	watchRequest.Key = model.ServiceKey{
		Namespace: getNamespace(pr.options),
		Service:   pr.target.URL.Host,
	}
	resp, err := consumerAPI.WatchService(watchRequest)
	if nil != err {
		return watchRequest.Key, nil, err
	}
	return watchRequest.Key, resp.EventChannel, nil
}

func (pr *polarisNamingResolver) watcher() {
	// defer pr.wg.Done()
	var consumerAPI api.ConsumerAPI
	var eventChan <-chan model.SubScribeEvent
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-pr.ctx.Done():
			return
		case <-pr.rn:
		case <-eventChan:
		case <-ticker.C:
		}
		var state *resolver.State
		var err error
		state, consumerAPI, err = pr.lookup()
		if err != nil {
			pr.cc.ReportError(err)
		} else {
			pr.balanceOnce.Do(func() {
				state.ServiceConfig = &serviceconfig.ParseResult{
					// lint:ignore SA1019 we want to keep the original config here
					Config: &grpc.ServiceConfig{
						LB: proto.String(scheme),
					},
				}
			})
			err = pr.cc.UpdateState(*state)
			if nil != err {
				log.Errorf("fail to do update service %s: %v", pr.target.URL.Host, err)
			}
			var svcKey model.ServiceKey
			svcKey, eventChan, err = pr.doWatch(consumerAPI)
			if nil != err {
				log.Errorf("fail to do watch for service %s: %v", svcKey, err)
			}
		}
	}
}

// Close resolver closed
func (pr *polarisNamingResolver) Close() {
	pr.cancel()
}
