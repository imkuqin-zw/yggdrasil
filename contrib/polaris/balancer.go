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

package polaris

import (
	"strings"
	"sync"
	"time"

	balancer2 "github.com/imkuqin-zw/yggdrasil/pkg/balancer"
	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	resolver2 "github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	polarisgo "github.com/polarismesh/polaris-go"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
)

func init() {
	balancer2.RegisterBuilder(name, newBalancer)
}

type balancer struct {
	dstServiceInfo *DstServiceInfo
	//srcServiceInfo *model.ServiceInfo
	routerAPI    polarisgo.RouterAPI
	consumerAPI  polarisgo.ConsumerAPI
	mu           sync.Mutex
	instances    *model.InstancesResponse
	instancesOld bool
}

func newBalancer(serviceName string) balancer2.Balancer {
	sdkCtx, err := Context()
	if err != nil {
		logger.ErrorField("fault to get polaris context", logger.Err(err))
	}

	br := &balancer{
		dstServiceInfo: getDstServiceInfo(serviceName),
		//srcServiceInfo: buildSourceInfo(),
		routerAPI:    polarisgo.NewRouterAPIByContext(sdkCtx),
		consumerAPI:  polarisgo.NewConsumerAPIByContext(sdkCtx),
		instancesOld: true,
	}
	return br
}

func (b *balancer) getInstances() (*model.InstancesResponse, error) {
	info := b.dstServiceInfo
	instancesRequest := &polarisgo.GetInstancesRequest{}
	instancesRequest.Namespace = info.Namespace
	instancesRequest.Service = info.ServiceName
	instancesRequest.SourceService = buildSourceInfo()
	if len(info.DstMetadata) > 0 {
		instancesRequest.Metadata = info.DstMetadata
	}
	instancesRequest.SkipRouteFilter = true
	resp, err := b.consumerAPI.GetInstances(instancesRequest)
	if nil != err {
		return nil, err
	}
	var instances = make([]model.Instance, 0, len(resp.Instances))
	totalWeight := 0
	for _, instance := range resp.Instances {
		if !instance.IsHealthy() || instance.IsIsolated() {
			continue
		}
		totalWeight += instance.GetWeight()
		instances = append(instances, instance)
	}
	resp.Instances = instances
	resp.TotalWeight = totalWeight
	return resp, nil
}

func (b *balancer) GetPicker() balancer2.Picker {
	return &picker{
		br:           b,
		routerAPI:    b.routerAPI,
		consumerAPI:  b.consumerAPI,
		onceExecute:  false,
		instances:    b.instances,
		instancesOld: b.instancesOld,
	}
}

func (b *balancer) Update(_ config2.Values) {
	resp, err := b.getInstances()
	if err != nil {
		logger.ErrorField("fault to get instances", logger.Err(err))
		b.instancesOld = true
		return
	}
	b.instancesOld = false
	b.instances = resp
	return
}

func (b *balancer) Close() error {
	b.consumerAPI.Destroy()
	return nil
}

func (b *balancer) Name() string {
	return name
}

type pickResult struct {
	endpoint endpoint
	report   func(err error)
}

func (p *pickResult) Endpoint() resolver2.Endpoint {
	return p.endpoint
}

func (p *pickResult) Report(err error) {
	p.report(err)
}

type picker struct {
	instancesOld        bool
	br                  *balancer
	consumerAPI         polarisgo.ConsumerAPI
	routerAPI           polarisgo.RouterAPI
	onceExecute         bool
	routerInstancesResp *model.InstancesResponse
	instances           *model.InstancesResponse
}

func (p *picker) Next(ri balancer2.RpcInfo) (balancer2.PickResult, error) {
	if !p.onceExecute {
		if p.instancesOld {
			resp, err := p.br.getInstances()
			if err != nil {
				return nil, err
			}
			p.instancesOld = true
			p.instances = resp
		}
		p.onceExecute = true
		routerRequest := &polarisgo.ProcessRoutersRequest{}
		routerRequest.DstInstances = p.instances
		routerRequest.SourceService = *buildSourceInfo()
		if md, ok := metadata.FromOutContext(ri.Ctx); ok {
			for k, v := range md {
				routerRequest.AddArguments(model.BuildCustomArgument(k, strings.Join(v, ";")))
				//routerRequest.SourceService.Metadata[k] = strings.Join(v, ";")
			}
		}
		routerRequest.Method = ri.Method
		var routerInstancesResp *model.InstancesResponse
		var err error
		routerInstancesResp, err = p.routerAPI.ProcessRouters(routerRequest)
		if nil != err {
			if sdkErr, ok := err.(model.SDKError); ok && sdkErr.ErrorCode() == model.ErrCodeRouteRuleNotMatch {
				routerInstancesResp = p.instances
			}
			logger.ErrorField("fault to process router", logger.Err(err))
			return nil, err
		}
		p.routerInstancesResp = routerInstancesResp
	}
	if p.routerInstancesResp == nil || len(p.routerInstancesResp.GetInstances()) == 0 {
		return nil, balancer2.ErrNoAvailableInstance
	}
	lbReq := &polarisgo.ProcessLoadBalanceRequest{}
	lbReq.DstInstances = p.routerInstancesResp
	lbReq.LbPolicy = config.DefaultLoadBalancerWR
	oneInsResp, err := p.routerAPI.ProcessLoadBalance(lbReq)
	if nil != err {
		return nil, err
	}
	instance := oneInsResp.GetInstance()
	if instance == nil {
		return nil, balancer2.ErrNoAvailableInstance
	}
	rp := &resultReporter{
		fullMethod:  ri.Method,
		instance:    instance,
		consumerAPI: p.br.consumerAPI,
		startTime:   time.Now(),
	}
	result := &pickResult{
		endpoint: instanceToEndpoint(instance),
		report:   rp.report,
	}
	return result, nil
}

type resultReporter struct {
	fullMethod  string
	instance    model.Instance
	consumerAPI polarisgo.ConsumerAPI
	startTime   time.Time
}

func (r *resultReporter) report(err error) {
	st := status.FromError(err)
	callResult := &polarisgo.ServiceCallResult{}
	callResult.CalledInstance = r.instance
	callResult.SourceService = buildSourceInfo()
	callResult.Method = r.fullMethod
	if st.Err() != nil {
		callResult.RetStatus = api.RetFail
	} else {
		callResult.RetStatus = api.RetSuccess
	}

	callResult.SetDelay(time.Since(r.startTime))
	callResult.SetRetCode(st.Code())
	if err := r.consumerAPI.UpdateServiceCallResult(callResult); err != nil {
		logger.ErrorField("polaris fault to report call info", logger.Err(err))
	}
}
