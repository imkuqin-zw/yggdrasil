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

package flow

import (
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/flow/configuration"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/configconnector"
	"github.com/modern-go/reflect2"

	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/flow/cbcheck"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/flow/data"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/flow/quota"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/flow/schedule"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model/pb"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/circuitbreaker"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/common"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/loadbalancer"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/localregistry"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/serverconnector"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/servicerouter"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/statreporter"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/plugin/subscribe"
)

// Engine ?????????????????????API???????????????????????????
type Engine struct {
	// ??????????????????
	connector serverconnector.ServerConnector
	// ??????????????????
	configConnector configconnector.ConfigConnector
	// ??????????????????
	registry localregistry.LocalRegistry
	// ????????????
	configuration config.Configuration
	// ???????????????????????????????????????
	filterOnlyRouter servicerouter.ServiceRouter
	// ?????????????????????
	routerChain *servicerouter.RouterChain
	// ???????????????
	reporterChain []statreporter.StatReporter
	// ???????????????
	loadbalancer loadbalancer.LoadBalancer
	// ???????????????????????????
	flowQuotaAssistant *quota.FlowQuotaAssistant
	// ?????????????????????reportclient
	globalCtx model.ValueContext
	// ??????????????????
	serverServices config.ServerServices
	// ????????????
	plugins plugin.Supplier
	// ??????????????????
	taskRoutines []schedule.TaskRoutine
	// ????????????????????????
	rtCircuitBreakChan chan<- *model.PriorityTask
	// ??????????????????????????????
	circuitBreakTask *cbcheck.CircuitBreakCallBack
	// ???????????????
	circuitBreakerChain []circuitbreaker.InstanceCircuitBreaker
	// ???????????????????????????
	subscribe subscribe.Subscribe
	// ?????????????????????
	configFileService *configuration.ConfigFileService
}

// InitFlowEngine ?????????flowEngine??????
func InitFlowEngine(flowEngine *Engine, initContext plugin.InitContext) error {
	var err error
	cfg := initContext.Config
	plugins := initContext.Plugins
	globalCtx := initContext.ValueCtx
	flowEngine.configuration = cfg
	flowEngine.plugins = plugins
	// ????????????????????????
	flowEngine.connector, err = data.GetServerConnector(cfg, plugins)
	if err != nil {
		return err
	}
	flowEngine.serverServices = config.GetServerServices(cfg)
	// ????????????????????????
	flowEngine.registry, err = data.GetRegistry(cfg, plugins)
	if err != nil {
		return err
	}
	if cfg.GetGlobal().GetStatReporter().IsEnable() {
		flowEngine.reporterChain, err = data.GetStatReporterChain(cfg, plugins)
		if err != nil {
			return err
		}
	}

	// ???????????????????????????
	if len(cfg.GetConfigFile().GetConfigConnectorConfig().GetAddresses()) > 0 {
		flowEngine.configConnector, err = data.GetConfigConnector(cfg, plugins)
		if err != nil {
			return err
		}
	}

	// ???????????????????????????
	err = flowEngine.LoadFlowRouteChain()
	if err != nil {
		return err
	}
	// ?????????????????????
	flowEngine.flowQuotaAssistant = &quota.FlowQuotaAssistant{}
	if err = flowEngine.flowQuotaAssistant.Init(flowEngine, flowEngine.configuration, flowEngine.plugins); err != nil {
		return err
	}
	// ?????????????????????
	flowEngine.globalCtx = globalCtx
	// ??????????????????
	when := cfg.GetConsumer().GetHealthCheck().GetWhen()
	disableHealthCheck := when == config.HealthCheckNever
	if !disableHealthCheck {
		if err = flowEngine.addHealthCheckTask(); err != nil {
			return err
		}
	}
	// ?????????????????????
	enable := cfg.GetConsumer().GetCircuitBreaker().IsEnable()
	if enable {
		flowEngine.circuitBreakerChain, err = data.GetCircuitBreakers(cfg, plugins)
		if err != nil {
			return err
		}
		flowEngine.rtCircuitBreakChan, flowEngine.circuitBreakTask, err = flowEngine.addPeriodicCircuitBreakTask()
		if err != nil {
			return err
		}
	}
	// ????????????????????????
	pluginName := cfg.GetConsumer().GetSubScribe().GetType()
	p, err := flowEngine.plugins.GetPlugin(common.TypeSubScribe, pluginName)
	if err != nil {
		return err
	}
	sP := p.(subscribe.Subscribe)
	flowEngine.subscribe = sP
	callbackHandler := common.PluginEventHandler{
		Callback: flowEngine.ServiceEventCallback,
	}
	initContext.Plugins.RegisterEventSubscriber(common.OnServiceUpdated, callbackHandler)
	globalCtx.SetValue(model.ContextKeyEngine, flowEngine)

	//???????????????????????????
	if cfg.GetConfigFile().IsEnable() {
		flowEngine.configFileService = configuration.NewConfigFileService(flowEngine.configConnector, flowEngine.configuration)
	}

	return nil
}

// LoadFlowRouteChain ???????????????????????????
func (e *Engine) LoadFlowRouteChain() error {
	var err error
	e.routerChain, err = data.GetServiceRouterChain(e.configuration, e.plugins)
	if err != nil {
		return err
	}
	filterOnlyRouterPlugin, err := e.plugins.GetPlugin(common.TypeServiceRouter, config.DefaultServiceRouterFilterOnly)
	if err != nil {
		return err
	}
	e.filterOnlyRouter = filterOnlyRouterPlugin.(servicerouter.ServiceRouter)
	// ????????????????????????
	e.loadbalancer, err = data.GetLoadBalancer(e.configuration, e.plugins)
	if err != nil {
		return err
	}
	return nil
}

// FlowQuotaAssistant ?????????????????????
func (e *Engine) FlowQuotaAssistant() *quota.FlowQuotaAssistant {
	return e.flowQuotaAssistant
}

// PluginSupplier ??????????????????
func (e *Engine) PluginSupplier() plugin.Supplier {
	return e.plugins
}

// WatchService watch service
func (e *Engine) WatchService(req *model.WatchServiceRequest) (*model.WatchServiceResponse, error) {
	if e.subscribe != nil {
		allInsReq := &model.GetAllInstancesRequest{}
		allInsReq.Namespace = req.Key.Namespace
		allInsReq.Service = req.Key.Service
		allInsRsp, err := e.SyncGetAllInstances(allInsReq)
		if err != nil {
			return nil, err
		}
		v, err := e.subscribe.WatchService(req.Key)
		if err != nil {
			log.GetBaseLogger().Errorf("watch service %s %s error:%s", req.Key.Namespace, req.Key.Service,
				err.Error())
			return nil, err
		}
		svcEventKey := &model.ServiceEventKey{
			ServiceKey: req.Key,
			Type:       model.EventInstances,
		}
		err = e.registry.WatchService(svcEventKey)
		if err != nil {
			return nil, err
		}
		watchResp := &model.WatchServiceResponse{}
		if e.subscribe.Name() == config.SubscribeLocalChannel {
			watchResp.EventChannel = v.(chan model.SubScribeEvent)
		} else {
			watchResp.EventChannel = nil
		}
		watchResp.GetAllInstancesResp = allInsRsp
		return watchResp, nil
	} else {
		return nil, model.NewSDKError(model.ErrCodeInternalError, nil, "engine subscribe is nil")
	}
}

func (e *Engine) GetContext() model.ValueContext {
	return e.globalCtx
}

// ServiceEventCallback serviceUpdate??????????????????
func (e *Engine) ServiceEventCallback(event *common.PluginEvent) error {
	if e.subscribe != nil {
		err := e.subscribe.DoSubScribe(event)
		if err != nil {
			log.GetBaseLogger().Errorf("subscribePlugin.DoSubScribe name:%s error:%s",
				e.subscribe.Name(), err.Error())
		}
	}
	return nil
}

// Start ????????????
func (e *Engine) Start() error {
	// ?????????????????????????????????
	clientReportTaskValues, err := e.addClientReportTask()
	if err != nil {
		return err
	}
	// ?????????????????????????????????
	serverServiceTaskValues, err := e.addLoadServerServiceTask()
	if err != nil {
		return err
	}
	// ????????????sdk????????????
	configReportTaskValues := e.addSDKConfigReportTask()
	// ????????????
	discoverSvc := e.serverServices.GetClusterService(config.DiscoverCluster)
	if nil != discoverSvc {
		schedule.StartTask(
			taskServerService, serverServiceTaskValues, map[interface{}]model.TaskValue{
				keyDiscoverService: &data.ServiceKeyComparable{SvcKey: discoverSvc.ServiceKey}})
	}
	schedule.StartTask(
		taskClientReport, clientReportTaskValues, map[interface{}]model.TaskValue{
			taskClientReport: &data.AllEqualsComparable{}})
	schedule.StartTask(
		taskConfigReport, configReportTaskValues, map[interface{}]model.TaskValue{
			taskConfigReport: &data.AllEqualsComparable{}})
	return nil
}

// getRouterChain ???????????????????????????
func (e *Engine) getRouterChain(svcInstances model.ServiceInstances) *servicerouter.RouterChain {
	svcInstancesProto := svcInstances.(*pb.ServiceInstancesInProto)
	routerChain := svcInstancesProto.GetServiceRouterChain()
	if nil == routerChain {
		return e.routerChain
	}
	return routerChain
}

// getLoadBalancer ?????????????????????????????????
// ?????????????????????????????????????????????????????????????????????????????????
func (e *Engine) getLoadBalancer(svcInstances model.ServiceInstances, chooseAlgorithm string) (
	loadbalancer.LoadBalancer, error) {
	svcInstancesProto := svcInstances.(*pb.ServiceInstancesInProto)
	svcLoadbalancer := svcInstancesProto.GetServiceLoadbalancer()
	if reflect2.IsNil(svcLoadbalancer) {
		if chooseAlgorithm == "" {
			return e.loadbalancer, nil
		} else {
			return data.GetLoadBalancerByLbType(chooseAlgorithm, e.plugins)
		}
	}
	return svcLoadbalancer, nil
}

// Destroy ??????????????????
func (e *Engine) Destroy() error {
	if len(e.taskRoutines) > 0 {
		for _, routine := range e.taskRoutines {
			routine.Destroy()
		}
	}
	if e.flowQuotaAssistant != nil {
		e.flowQuotaAssistant.Destroy()
	}
	if e.configFileService != nil {
		e.configFileService.Destroy()
	}
	return nil
}

// SyncReportStat ????????????????????????????????????
func (e *Engine) SyncReportStat(typ model.MetricType, stat model.InstanceGauge) error {
	if !model.ValidMetircType(typ) {
		return model.NewSDKError(model.ErrCodeAPIInvalidArgument, nil, "invalid report metric type")
	}
	if len(e.reporterChain) > 0 {
		for _, reporter := range e.reporterChain {
			err := reporter.ReportStat(typ, stat)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// reportAPIStat ??????api??????
func (e *Engine) reportAPIStat(result *model.APICallResult) error {
	return e.SyncReportStat(model.SDKAPIStat, result)
}

// reportSvcStat ??????????????????
func (e *Engine) reportSvcStat(result *model.ServiceCallResult) error {
	return e.SyncReportStat(model.ServiceStat, result)
}
