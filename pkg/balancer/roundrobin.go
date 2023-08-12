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

package balancer

import (
	"sync/atomic"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"google.golang.org/genproto/googleapis/rpc/code"
)

func init() {
	RegisterBuilder("round_robin", newRoundRobin)
}

type instance struct {
	Address  string                 `yaml:"address"`
	Protocol string                 `yaml:"protocol"`
	Metadata map[string]interface{} `yaml:"metadata"`
}

func (i *instance) GetAddress() string {
	return i.Address
}

func (i *instance) GetProtocol() string {
	return i.Protocol
}

func (i *instance) GetMetadata() map[string]interface{} {
	return i.Metadata
}

type pickResult struct {
	endpoint *instance
}

func (p *pickResult) Endpoint() resolver.Endpoint {
	return p.endpoint
}

func (p *pickResult) Report(error) {
	return
}

type roundRobinPicker struct {
	idx      int64
	endpoint []*instance
}

func (r *roundRobinPicker) Next(RpcInfo) (PickResult, error) {
	endpoints := r.endpoint
	if len(endpoints) == 0 {
		return nil, status.Errorf(code.Code_UNAVAILABLE, "not found endpoint")
	}
	idx := r.idx % int64(len(r.endpoint))
	res := &pickResult{endpoint: r.endpoint[idx]}
	r.idx++
	return res, nil
}

type RoundRobin struct {
	idx      atomic.Int64
	endpoint []*instance
}

func newRoundRobin(string) Balancer {
	return &RoundRobin{}
}

func (b *RoundRobin) GetPicker() Picker {
	return &roundRobinPicker{
		idx:      b.idx.Add(1),
		endpoint: b.endpoint,
	}
}

func (b *RoundRobin) Update(values config.Values) {
	endpoints := make([]*instance, 0)
	if err := values.Get(config.KeySingleEndpoints).Scan(&endpoints); err != nil {
		logger.ErrorField("fault to load endpoints config", logger.Err(err))
		return
	}
	b.endpoint = endpoints
}

func (b *RoundRobin) Close() error {
	return nil
}

func (b *RoundRobin) Name() string {
	return "round_robin"
}
