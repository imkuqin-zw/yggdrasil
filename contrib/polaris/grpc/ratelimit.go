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

	"github.com/imkuqin-zw/yggdrasil/contrib/polaris"
	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	grpc2 "github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	grpc2.RegisterUnaryInterceptor("polaris_limit", newLimitUnaryInterceptor)
}

func newLimitUnaryInterceptor() grpc.UnaryServerInterceptor {
	polarisCtx, _ := polaris.Context()
	limitAPI := api.NewLimitAPIByContext(polarisCtx)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		quotaReq := api.NewQuotaRequest()
		namespace := pkg.Namespace()
		serviceName := pkg.Name()
		quotaReq.SetNamespace(namespace)
		quotaReq.SetService(serviceName)
		quotaReq.SetLabels(map[string]string{
			"scheme":  "grpc",
			"methods": info.FullMethod,
		})
		future, err := limitAPI.GetQuota(quotaReq)
		if nil != err {
			log.Errorf("fail to do ratelimit %s: %v", info.FullMethod, err)
			return handler(ctx, req)
		}
		rsp := future.Get()
		if rsp.Code == api.QuotaResultLimited {
			return nil, status.Error(codes.ResourceExhausted, rsp.Info)
		}
		return handler(ctx, req)
	}
}
