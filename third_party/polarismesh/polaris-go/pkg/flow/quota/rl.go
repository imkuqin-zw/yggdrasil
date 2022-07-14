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

package quota

import (
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model"
)

// RateLimitType 限流的类型
type RateLimitType int

const (
	TrafficShapingLimited RateLimitType = 0
	QuotaLimited          RateLimitType = 1
	WindowDeleted         RateLimitType = 2
	// QuotaRequested        RateLimitType = 3
	QuotaGranted RateLimitType = 4
)

type LimitMode int

const (
	// 未知类型，用于兼容前面pb
	LimitUnknownMode LimitMode = 0
	// 全局类型，与限流server发生交互
	LimitGlobalMode LimitMode = 1
	// 本地类型，使用本地限流算法
	LimitLocalMode LimitMode = 2
	// 降级类型，因为无法连接限流server，导致必须使用本地限流算法
	LimitDegradeMode LimitMode = 3
)

// 限流统计gauge
type RateLimitGauge struct {
	model.EmptyInstanceGauge
	Window    *RateLimitWindow
	Namespace string
	Service   string
	Type      RateLimitType
	Labels    map[string]string
	// 限流周期， 单位秒
	Duration uint32
	// 限流发生时的mode, 和plugin的pb要保持一致
	LimitModeType LimitMode
}
