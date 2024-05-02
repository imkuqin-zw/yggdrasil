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
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
)

type endpoint struct {
	Address  string
	Protocol string
	Metadata map[string]interface{}
}

func (e endpoint) GetAddress() string {
	return e.Address
}

func (e endpoint) GetProtocol() string {
	return e.Protocol
}

func (e endpoint) GetMetadata() map[string]interface{} {
	return e.Metadata
}

func instanceToEndpoint(i model.Instance) endpoint {
	e := endpoint{
		Address:  fmt.Sprintf("%s:%d", i.GetHost(), i.GetPort()),
		Protocol: i.GetProtocol(),
	}
	e.Metadata = map[string]interface{}{
		// 实例所在命名空间
		"namespace": i.GetNamespace(),
		// 服务实例唯一标识
		"id": i.GetId(),
		// 实例的vpcId
		"vpcId": i.GetVpcId(),
		// 实例版本号
		"version": i.GetVersion(),
		// 实例静态权重值
		"weight": i.GetWeight(),
		// 实例优先级信息
		"priority": i.GetPriority(),
		// 实例元数据信息
		"metadata": i.GetMetadata(),
		// 实例逻辑分区
		"logicSet": i.GetLogicSet(),
		// 实例是否健康，基于服务端返回的健康数据
		"isHealthy": i.IsHealthy(),
		// 实例是否已经被手动隔离
		"isIsolated": i.IsIsolated(),
		// 实例是否启动了健康检查
		"isEnableHealthCheck": i.IsEnableHealthCheck(),
		// 实例所属的大区信息
		"region": i.GetRegion(),
		// 实例所属的地方信息
		"zone": i.GetZone(),
		// 实例所属的园区信息
		"campus": i.GetCampus(),
	}
	return e
}

type DstServiceInfo struct {
	Namespace   string
	ServiceName string
	DstMetadata map[string]string
}

func getDstServiceInfo(serviceName string) *DstServiceInfo {
	return &DstServiceInfo{
		Namespace:   namespace(config2.Get(config2.KeyClientNamespace).String(pkg.Namespace())),
		ServiceName: serviceName,
		DstMetadata: config2.GetMulti(configKeyDst, configKeyClientDst).StringMap(map[string]string{}),
	}
}
