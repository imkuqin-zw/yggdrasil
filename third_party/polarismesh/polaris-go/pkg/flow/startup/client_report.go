/**
 * Tencent is pleased to support the open source community by making polaris-go available.
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

package startup

import (
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/flow/data"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model"
	namingpb "github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model/pb/v1"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/localregistry"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/reporthandler"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/serverconnector"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/version"
)

// NewReportClientCallBack  创建上报回调
func NewReportClientCallBack(
	cfg config.Configuration, supplier plugin.Supplier, globalCtx model.ValueContext) (*ReportClientCallBack, error) {
	var err error
	var callback = &ReportClientCallBack{}
	if callback.connector, err = data.GetServerConnector(cfg, supplier); err != nil {
		return nil, err
	}
	if callback.registry, err = data.GetRegistry(cfg, supplier); err != nil {
		return nil, err
	}
	if callback.reportChain, err = data.GetReportChain(cfg, supplier); err != nil {
		return nil, err
	}
	callback.configuration = cfg
	callback.globalCtx = globalCtx
	callback.interval = cfg.GetGlobal().GetAPI().GetReportInterval()
	callback.loadLocalClientReportResult()
	return callback, nil
}

// ReportClientCallBack 上报客户端状态任务回调
type ReportClientCallBack struct {
	connector     serverconnector.ServerConnector
	registry      localregistry.InstancesRegistry
	configuration config.Configuration
	globalCtx     model.ValueContext
	interval      time.Duration
	reportChain   *reporthandler.ReportHandlerChain
}

const (
	// 地域信息持久化数据
	clientInfoPersistFile = "client_info.json"
)

// loadLocalClientReportResult 从本地缓存加载上报结果信息
func (r *ReportClientCallBack) loadLocalClientReportResult() {
	resp := &namingpb.Response{}
	cachedFile := clientInfoPersistFile
	err := r.registry.LoadPersistedMessage(cachedFile, resp)
	if err != nil {
		log.GetBaseLogger().Warnf("fail to load local region info from %s, err is %v", cachedFile, err)
		return
	}

	client := resp.Client

	for i := range r.reportChain.Chain {
		handler := r.reportChain.Chain[i]
		handler.InitLocal(client)
	}
}

// reportClientRequest 客户端上报的请求
func (r *ReportClientCallBack) reportClientRequest() *model.ReportClientRequest {
	apiConfig := r.configuration.GetGlobal().GetAPI()
	clientHost := apiConfig.GetBindIP()
	reportClientReq := &model.ReportClientRequest{
		Version: version.Version,
		Timeout: r.configuration.GetGlobal().GetAPI().GetTimeout(),
		PersistHandler: func(message proto.Message) error {
			return r.registry.PersistMessage(clientInfoPersistFile, message)
		},
	}
	if len(clientHost) > 0 {
		reportClientReq.Host = clientHost
	}
	return reportClientReq
}

// Process 执行任务
func (r *ReportClientCallBack) Process(
	taskKey interface{}, taskValue interface{}, lastProcessTime time.Time) model.TaskResult {
	if !lastProcessTime.IsZero() && time.Since(lastProcessTime) < r.interval {
		return model.SKIP
	}
	reportClientReq := r.reportClientRequest()
	if err := reportClientReq.Validate(); err != nil {
		log.GetBaseLogger().Errorf("report client request fatal validate error:%v", err)
		return model.TERMINATE
	}

	for i := range r.reportChain.Chain {
		handler := r.reportChain.Chain[i]
		handler.HandleRequest(reportClientReq)
	}

	reportClientResp, err := r.connector.ReportClient(reportClientReq)
	if err != nil {
		log.GetBaseLogger().Errorf("report client info:%+v, error:%v", reportClientReq, err)
		// 发生错误也要重试，直到获取到地域信息为止
		for i := range r.reportChain.Chain {
			handler := r.reportChain.Chain[i]
			handler.HandleResponse(reportClientResp, err)
		}
		return model.CONTINUE
	}

	for i := range r.reportChain.Chain {
		handler := r.reportChain.Chain[i]
		handler.HandleResponse(reportClientResp, nil)
	}
	return model.CONTINUE
}

// OnTaskEvent 任务事件回调
func (r *ReportClientCallBack) OnTaskEvent(event model.TaskEvent) {

}
