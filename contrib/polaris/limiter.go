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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/interceptor"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/peer"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/flow/data"
	"github.com/polarismesh/polaris-go/pkg/model"
	v1 "github.com/polarismesh/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/genproto/googleapis/rpc/code"
)

func init() {
	interceptor.RegisterUnaryServerIntBuilder(fmt.Sprintf("%s_limit", name), newRateLimitUnaryServerInterceptor)
}

// rateLimitInterceptor is a gRPC interceptor that implements rate limiting.
type rateLimitInterceptor struct {
	srcServiceInfo *model.ServiceInfo
	limitAPI       api.LimitAPI
}

func newRateLimitUnaryServerInterceptor() interceptor.UnaryServerInterceptor {
	ri := &rateLimitInterceptor{}
	ri.srcServiceInfo = buildSourceInfo()
	polarisCtx, _ := Context()
	ri.limitAPI = api.NewLimitAPIByContext(polarisCtx)
	return ri.UnaryInterceptor
}

// UnaryInterceptor returns a unary interceptor for rate limiting.
func (p *rateLimitInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *interceptor.UnaryServerInfo, handler interceptor.UnaryHandler) (resp interface{}, err error) {
	quotaReq := p.buildQuotaRequest(ctx, req, info)
	if quotaReq == nil {
		return handler(ctx, req)
	}

	future, err := p.limitAPI.GetQuota(quotaReq)
	if nil != err {
		logger.Errorf("polaris rateLimit fail to get quota %#v: %v", quotaReq, err)
		return handler(ctx, req)
	}

	if rsp := future.Get(); rsp.Code == api.QuotaResultLimited {
		return nil, status.Errorf(code.Code_RESOURCE_EXHAUSTED, rsp.Info)
	}
	return handler(ctx, req)
}

func (p *rateLimitInterceptor) buildQuotaRequest(ctx context.Context, _ interface{}, info *interceptor.UnaryServerInfo) api.QuotaRequest {
	fullMethodName := info.FullMethod
	tokens := strings.Split(fullMethodName, "/")
	if len(tokens) != 3 {
		return nil
	}
	quotaReq := api.NewQuotaRequest()
	quotaReq.SetNamespace(p.srcServiceInfo.Namespace)
	quotaReq.SetService(p.srcServiceInfo.Service)
	quotaReq.SetMethod(fullMethodName)
	if len(p.srcServiceInfo.Service) > 0 {
		quotaReq.SetService(p.srcServiceInfo.Service)
		quotaReq.SetMethod(fullMethodName)
	}

	matchs, ok := p.fetchArguments(quotaReq.(*model.QuotaRequestImpl))
	if !ok {
		return quotaReq
	}
	header, ok := metadata.FromInContext(ctx)
	if !ok {
		header = metadata.MD{}
	}

	for i := range matchs {
		item := matchs[i]
		switch item.GetType() {
		case v1.MatchArgument_CALLER_SERVICE:
			serviceValues := header.Get(polarisCallerServiceKey)
			namespaceValues := header.Get(polarisCallerNamespaceKey)
			if len(serviceValues) > 0 && len(namespaceValues) > 0 {
				quotaReq.AddArgument(model.BuildCallerServiceArgument(namespaceValues[0], serviceValues[0]))
			}
		case v1.MatchArgument_HEADER:
			values := header.Get(item.GetKey())
			if len(values) > 0 {
				quotaReq.AddArgument(model.BuildHeaderArgument(item.GetKey(), fmt.Sprintf("%+v", values[0])))
			}
		case v1.MatchArgument_CALLER_IP:
			if pr, ok := peer.PeerFromContext(ctx); ok && pr.Addr != nil {
				address := pr.Addr.String()
				addrSlice := strings.Split(address, ":")
				if len(addrSlice) == 2 {
					clientIP := addrSlice[0]
					quotaReq.AddArgument(model.BuildCallerIPArgument(clientIP))
				}
			}
		}
	}

	return quotaReq
}

func (p *rateLimitInterceptor) fetchArguments(req *model.QuotaRequestImpl) ([]*v1.MatchArgument, bool) {
	engine := p.limitAPI.SDKContext().GetEngine()

	getRuleReq := &data.CommonRateLimitRequest{
		DstService: model.ServiceKey{
			Namespace: req.GetNamespace(),
			Service:   req.GetService(),
		},
		Trigger: model.NotifyTrigger{
			EnableDstRateLimit: true,
		},
		ControlParam: model.ControlParam{
			Timeout: time.Millisecond * 500,
		},
	}

	if err := engine.SyncGetResources(getRuleReq); err != nil {
		logger.Errorf("polaris rateLimit ns:%s svc:%s get RateLimit Rule fail: %+v",
			req.GetNamespace(), req.GetService(), err)
		return nil, false
	}

	svcRule := getRuleReq.RateLimitRule
	if svcRule == nil || svcRule.GetValue() == nil {
		logger.Warnf("polaris rateLimit ns:%s svc:%s get RateLimit Rule is nil",
			req.GetNamespace(), req.GetService())
		return nil, false
	}

	rules, ok := svcRule.GetValue().(*v1.RateLimit)
	if !ok {
		logger.Errorf("polaris rateLimit ns:%s svc:%s get RateLimit Rule invalid",
			req.GetNamespace(), req.GetService())
		return nil, false
	}

	ret := make([]*v1.MatchArgument, 0, 4)
	for i := range rules.GetRules() {
		rule := rules.GetRules()[i]
		if len(rule.GetArguments()) == 0 {
			continue
		}

		ret = append(ret, rule.Arguments...)
	}
	return ret, true
}
