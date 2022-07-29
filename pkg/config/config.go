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

// Package conf is an interface for dynamic configuration.
package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
	"go.uber.org/atomic"
)

type config struct {
	keyDelimiter string

	sourceDataMu sync.RWMutex
	sourceData   [types.ConfigPriorityMax]map[string]interface{}

	cacheMu sync.Mutex
	kvs     map[string]interface{}
	vs      atomic.Value

	watcherMu sync.RWMutex
	watchers  map[string][]func(event types.ConfigWatchEvent)
}

func (c *config) Map() map[string]interface{} {
	return c.vs.Load().(types.ConfigValues).Map()
}

func (c *config) Scan(v interface{}) error {
	return c.vs.Load().(types.ConfigValues).Scan(v)
}

func (c *config) Close() error {

	return nil
}

func (c *config) Get(key string) types.ConfigValue {
	return c.vs.Load().(types.ConfigValues).Get(key)
}

func (c *config) Set(key string, val interface{}) error {
	c.sourceDataMu.Lock()
	xmap.MergeStringMap(c.sourceData[types.ConfigPriorityMemory], c.buildSetMap(key, val))
	c.sourceDataMu.Unlock()
	if err := c.apply(); err != nil {
		return err
	}
	return nil
}

func (c *config) buildSetMap(key string, val interface{}) map[string]interface{} {
	keys := strings.Split(strings.ToLower(key), c.keyDelimiter)
	xarray.ReverseStringArray(keys)
	tmp := make(map[string]interface{})
	for i, k := range keys {
		if i == 0 {
			tmp[k] = val
			continue
		}
		tmp = map[string]interface{}{k: tmp}
	}
	return tmp
}

func (c *config) Del(key string) error {
	c.sourceDataMu.Lock()
	val := &values{val: c.sourceData[types.ConfigPriorityMemory]}
	if err := val.Del(key); err != nil {
		c.sourceDataMu.Unlock()
		return err
	}
	c.sourceData[types.ConfigPriorityMemory] = val.Map()
	c.sourceDataMu.Unlock()
	if err := c.apply(); err != nil {
		return err
	}
	return nil
}

func (c *config) Bytes() []byte {
	return c.vs.Load().(types.ConfigValues).Bytes()
}

func (c *config) AddWatcher(key string, watcher func(types.ConfigWatchEvent)) error {
	c.watcherMu.Lock()
	defer c.watcherMu.Unlock()
	c.watchers[key] = append(c.watchers[key], watcher)
	return nil
}

func (c *config) DelWatcher(key string, watcher func(types.ConfigWatchEvent)) error {
	c.watcherMu.Lock()
	defer c.watcherMu.Unlock()
	delete(c.watchers, key)
	return nil
}

func (c *config) LoadSource(sources ...types.ConfigSource) error {
	if err := c.loadSource(sources...); err != nil {
		return err
	}
	if err := c.apply(); err != nil {
		return err
	}
	return nil
}

func (c *config) watchSource(source types.ConfigSource) {
	changeCh, err := source.Watch()
	if err != nil {
		log.Errorf("fault to watch config source, err: %+v", err)
		return
	}
	for {
		select {
		case change, ok := <-changeCh:
			if !ok {
				return
			}
			if err := c.loadSourceData(change); err != nil {
				log.Errorf("fault to load source data, err: %+v", err)
			}
		}
	}

}

func (c *config) loadSourceData(sourceData types.ConfigSourceData) error {
	v := make(map[string]interface{})
	if err := sourceData.Unmarshal(&v); err != nil {
		return err
	}
	xmap.CoverInterfaceToStringMap(v)
	xmap.MergeStringMap(c.sourceData[sourceData.Priority()], v)
	return nil
}

func (c *config) loadSource(sources ...types.ConfigSource) error {
	c.sourceDataMu.Lock()
	defer c.sourceDataMu.Unlock()
	for _, item := range sources {
		sourceData, err := item.Read()
		if err != nil {
			return err
		}
		if err := c.loadSourceData(sourceData); err != nil {
			return err
		}
		if item.Changeable() {
			xgo.Go(func() { c.watchSource(item) }, nil)
		}
	}
	return nil
}

func (c *config) apply() error {
	c.sourceDataMu.RLock()
	override := make(map[string]interface{})
	xmap.MergeStringMap(override, c.sourceData[:]...)
	c.sourceDataMu.RUnlock()
	c.cacheMu.Lock()
	var changes = make(map[string]types.WatchEventType)
	kvs := c.traverse(override, c.keyDelimiter)
	for k, v := range kvs {
		orig, ok := c.kvs[k]
		if !ok {
			changes[k] = types.WatchEventAdd
		} else if !reflect.DeepEqual(orig, v) {
			changes[k] = types.WatchEventDel
		}
	}
	for k := range c.kvs {
		if _, ok := kvs[k]; !ok {
			changes[k] = types.WatchEventDel
		}
	}
	c.kvs = kvs
	c.vs.Store(newValues(c.keyDelimiter, override))
	c.cacheMu.Unlock()
	if len(changes) > 0 {
		c.notify(changes, &values{keyDelimiter: c.keyDelimiter, val: override})
	}
	return nil
}

func lookup(prefix string, target map[string]interface{}, data map[string]interface{}, sep string) {
	for k, v := range target {
		pp := fmt.Sprintf("%s%s%s", prefix, sep, k)
		if prefix == "" {
			pp = k
		}
		if dd, ok := v.(map[string]interface{}); ok {
			lookup(pp, dd, data, sep)
		} else {
			data[pp] = v
		}
	}
}

func (c *config) traverse(override map[string]interface{}, sep string) map[string]interface{} {
	data := make(map[string]interface{})
	lookup("", override, data, sep)
	return data
}

func (c *config) notify(changes map[string]types.WatchEventType, val types.ConfigValues) {
	c.watcherMu.RLock()
	defer c.watcherMu.RUnlock()
	var changedWatchPrefixMap = map[string]types.WatchEventType{}
	for watchPrefix := range c.watchers {
		for key, et := range changes {
			// 前缀匹配
			if c.hasPrefix(key, watchPrefix) {
				changedWatchPrefixMap[watchPrefix] = et
			}
		}
	}

	for changedWatchPrefix, et := range changedWatchPrefixMap {
		v := val.Get(changedWatchPrefix)
		for _, handle := range c.watchers[changedWatchPrefix] {
			xgo.Go(func() {
				handle(newConfigWatchEvent(et, v))
			}, nil)
		}
	}
}

func (c *config) hasPrefix(key, watchPrefix string) bool {
	return xstrings.HasPrefix(key, watchPrefix, c.keyDelimiter)
}

type watchEvent struct {
	cate  types.WatchEventType
	value types.ConfigValue
}

func (e *watchEvent) Type() types.WatchEventType {
	return e.cate
}

func (e *watchEvent) Value() types.ConfigValue {
	return e.value
}

func newConfigWatchEvent(cate types.WatchEventType, val types.ConfigValue) types.ConfigWatchEvent {
	return &watchEvent{cate: cate, value: val}
}

func NewConfig(keyDelimiter string) types.Config {
	cfg := &config{keyDelimiter: keyDelimiter}
	cfg.vs.Store(&values{keyDelimiter: keyDelimiter})
	for i := range cfg.sourceData {
		cfg.sourceData[i] = map[string]interface{}{}
	}
	return cfg
}
