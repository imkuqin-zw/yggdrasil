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
	"errors"
	"fmt"
	"sync"
	"time"

	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	resolver2 "github.com/imkuqin-zw/yggdrasil/pkg/resolver"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"
)

func init() {
	resolver2.RegisterBuilder(name, newResolver)
}

type resolver struct {
	wg         sync.WaitGroup
	sourceInfo *model.ServiceInfo
	mu         sync.RWMutex
	closed     bool
	watcher    map[string]*watcherInstance
}

func newResolver(_ string) (resolver2.Resolver, error) {
	rs := &resolver{watcher: map[string]*watcherInstance{}}
	rs.sourceInfo = buildSourceInfo()
	return rs, nil
}

func (pr *resolver) AddWatch(serviceName string) (err error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if pr.closed {
		return errors.New("resolver closed")
	}
	if _, ok := pr.watcher[serviceName]; ok {
		return nil
	}
	watcher := &watcherInstance{
		pr: pr,
	}
	watcher.info = getDstServiceInfo(serviceName)
	sdkCtx, err := Context()
	if nil != err {
		return err
	}
	watcher.consumer = api.NewConsumerAPIByContext(sdkCtx)
	pr.watcher[serviceName] = watcher
	xgo.Go(func() {
		watcher.watch()
	}, nil)
	return nil
}

func (pr *resolver) DelWatch(serviceName string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if pr.closed {
		return nil
	}
	w, ok := pr.watcher[serviceName]
	if !ok {
		return nil
	}
	w.stop()
	return nil
}

func (pr *resolver) Close() error {
	pr.mu.Lock()
	if pr.closed {
		return nil
	}
	pr.closed = true
	pr.mu.Unlock()
	for _, item := range pr.watcher {
		item.stop()
	}
	pr.wg.Wait()
	return nil
}

func (pr *resolver) Name() string {
	return name
}

type watcherInstance struct {
	info        *DstServiceInfo
	pr          *resolver
	consumer    api.ConsumerAPI
	watchCancel func()
}

func (w *watcherInstance) doWatch() error {
	watchRequest := &api.WatchServiceRequest{}
	watchRequest.Key = model.ServiceKey{
		Namespace: w.info.Namespace,
		Service:   w.info.ServiceName,
	}
	res, err := w.consumer.GetAllInstances(&api.GetAllInstancesRequest{
		GetAllInstancesRequest: model.GetAllInstancesRequest{
			Service:   w.info.ServiceName,
			Namespace: w.info.Namespace,
		},
	})
	if err != nil {
		return err
	}
	w.OnInstancesUpdate(res)
	resp, err := w.consumer.WatchAllInstances(&api.WatchAllInstancesRequest{
		WatchAllInstancesRequest: model.WatchAllInstancesRequest{
			ServiceKey: model.ServiceKey{
				Namespace: w.info.Namespace,
				Service:   w.info.ServiceName,
			},
			WatchMode:         api.WatchModeNotify,
			InstancesListener: w,
		},
	})
	if nil != err {
		logger.ErrorField("fail to do watch", logger.String("serviceName", w.info.ServiceName), logger.Err(err))
		return err
	}
	w.watchCancel = resp.CancelWatch
	return nil
}

func (w *watcherInstance) OnInstancesUpdate(resp *model.InstancesResponse) {
	var endpoints = make([]interface{}, 0, len(resp.Instances))
	for _, instance := range resp.Instances {
		if !instance.IsHealthy() || instance.IsIsolated() {
			continue
		}
		endpoint := map[string]interface{}{
			config2.KeySingleAddress:  fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
			config2.KeySingleProtocol: instance.GetProtocol(),
			config2.KeySingleMetadata: instance.GetMetadata(),
		}
		endpoints = append(endpoints, endpoint)
	}
	_ = config2.SetMulti([]string{fmt.Sprintf(config2.KeyClientEndpoints, w.info.ServiceName)}, []interface{}{endpoints})
}

func (w *watcherInstance) watch() {
	for {
		if err := w.doWatch(); err == nil {
			return
		}
		time.Sleep(time.Second * 5)
	}
}

func (w *watcherInstance) stop() {
	w.watchCancel()
	w.consumer.Destroy()
}
