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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/imkuqin-zw/yggdrasil/pkg/config/source"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
)

type versionValues struct {
	values  Values
	version uint64
}

type config struct {
	keyDelimiter string

	sourceDataMu sync.RWMutex
	sourceData   [source.PriorityMax]map[string]interface{}

	cacheMu sync.Mutex
	kvs     map[string]interface{}
	version uint64
	vs      atomic.Value

	watcherMu sync.RWMutex
	watchers  map[string][]func(event WatchEvent)
}

func (c *config) Map() map[string]interface{} {
	return c.vs.Load().(versionValues).values.Map()
}

func (c *config) Scan(v interface{}) error {
	return c.vs.Load().(versionValues).values.Scan(v)
}

func (c *config) Close() error {
	return nil
}

func (c *config) Get(key string) Value {
	return c.vs.Load().(versionValues).values.Get(key)
}

func (c *config) GetMulti(keys ...string) Value {
	return c.vs.Load().(versionValues).values.GetMulti(keys...)
}

func (c *config) ValueToValues(v Value) Values {
	return newValues(keyDelimiter, v.Map())
}

func (c *config) Set(key string, val interface{}) error {
	c.sourceDataMu.Lock()
	xmap.MergeStringMap(c.sourceData[source.PriorityMemory], c.buildSetMap(key, val))
	c.sourceDataMu.Unlock()
	if err := c.apply(); err != nil {
		return err
	}
	return nil
}

func (c *config) SetMulti(keys []string, values []interface{}) error {
	if len(keys) != len(values) {
		return errors.New("the quantity of key and value does not match")
	}
	c.sourceDataMu.Lock()
	for i := 0; i < len(keys); i++ {
		xmap.MergeStringMap(c.sourceData[source.PriorityMemory], c.buildSetMap(keys[i], values[i]))
	}
	c.sourceDataMu.Unlock()
	if err := c.apply(); err != nil {
		return err
	}
	return nil
}

func (c *config) toSetInterface(val interface{}) interface{} {
	switch v := val.(type) {
	case map[interface{}]interface{}:
		val = xmap.ToMapStringInterface(v)
	case map[string]interface{}:
		xmap.CoverInterfaceMapToStringMap(v)
	case []interface{}:
		for i, item := range v {
			v[i] = c.toSetInterface(item)
		}
	default:
	}
	return val
}

func (c *config) buildSetMap(key string, val interface{}) map[string]interface{} {
	val = c.toSetInterface(val)
	keys := genPath(strings.ToLower(key), c.keyDelimiter)
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
	val := &values{val: c.sourceData[source.PriorityMemory]}
	if err := val.Del(key); err != nil {
		c.sourceDataMu.Unlock()
		return err
	}
	c.sourceData[source.PriorityMemory] = val.Map()
	c.sourceDataMu.Unlock()
	if err := c.apply(); err != nil {
		return err
	}
	return nil
}

func (c *config) Bytes() []byte {
	return c.vs.Load().(versionValues).values.Bytes()
}

func (c *config) addWatcher(key string, watcher func(WatchEvent)) {
	c.watcherMu.Lock()
	defer c.watcherMu.Unlock()
	c.watchers[key] = append(c.watchers[key], watcher)
}

func (c *config) AddWatcher(key string, watcher func(WatchEvent)) error {
	c.addWatcher(key, watcher)
	vs := c.vs.Load().(versionValues)
	v := vs.values.Get(key)
	xgo.Go(func() {
		watcher(newConfigWatchEvent(WatchEventUpd, vs.version, v))
	}, nil)
	return nil
}

func (c *config) DelWatcher(key string, _ func(WatchEvent)) error {
	c.watcherMu.Lock()
	defer c.watcherMu.Unlock()
	delete(c.watchers, key)
	return nil
}

func (c *config) LoadSource(sources ...source.Source) error {
	if err := c.loadSource(sources...); err != nil {
		return err
	}
	if err := c.apply(); err != nil {
		return err
	}
	return nil
}

func (c *config) watchSource(source source.Source) {
	changeCh, err := source.Watch()
	if err != nil {
		logger.Errorf("fault to watch config source, err: %+v", err)
		return
	}
	for {
		select {
		case change, ok := <-changeCh:
			if !ok {
				return
			}
			if err := c.loadSourceData(change); err != nil {
				logger.Errorf("fault to load source data, err: %+v", err)
			}
		}
	}

}

func (c *config) loadSourceData(sourceData source.SourceData) error {
	v := make(map[string]interface{})
	if err := sourceData.Unmarshal(&v); err != nil {
		return err
	}
	xmap.CoverInterfaceMapToStringMap(v)
	xmap.MergeStringMap(c.sourceData[sourceData.Priority()], v)
	return nil
}

func (c *config) loadSource(sources ...source.Source) error {
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
	c.sourceDataMu.Lock()
	defer c.sourceDataMu.Unlock()
	override := make(map[string]interface{})
	xmap.MergeStringMap(override, c.sourceData[:]...)
	//c.cacheMu.Lock()
	var (
		version uint64
		changes = make(map[string]WatchEventType)
	)

	kvs := c.traverse(override, c.keyDelimiter)
	for k, v := range kvs {
		orig, ok := c.kvs[k]
		if !ok {
			changes[k] = WatchEventAdd
		} else if !reflect.DeepEqual(orig, v) {
			changes[k] = WatchEventUpd
		}
	}
	for k := range c.kvs {
		if _, ok := kvs[k]; !ok {
			changes[k] = WatchEventDel
		}
	}
	if len(changes) == 0 {
		//c.cacheMu.Unlock()
		return nil
	}
	c.kvs = kvs
	c.version++
	version = c.version
	c.vs.Store(versionValues{version: version, values: newValues(c.keyDelimiter, override)})
	//c.cacheMu.Unlock()
	if len(changes) > 0 {
		c.notify(changes, version, newValues(c.keyDelimiter, override))
	}
	return nil
}

func lookup(prefix string, target map[string]interface{}, data map[string]interface{}, sep string) {
	for k, v := range target {
		if strings.Index(k, ".") > 0 {
			k = fmt.Sprintf("{%s}", k)
		}
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

func (c *config) notify(changes map[string]WatchEventType, version uint64, val Values) {
	c.watcherMu.RLock()
	defer c.watcherMu.RUnlock()
	var changedWatchPrefixMap = map[string]WatchEventType{}
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
				handle(newConfigWatchEvent(et, version, v))
			}, nil)
		}
	}
}

func (c *config) hasPrefix(key, watchPrefix string) bool {
	return xstrings.HasPrefix(key, watchPrefix, c.keyDelimiter)
}

type watchEvent struct {
	cate    WatchEventType
	value   Value
	version uint64
}

func (e *watchEvent) Type() WatchEventType {
	return e.cate
}

func (e *watchEvent) Value() Value {
	return e.value
}

func (e *watchEvent) Version() uint64 {
	return e.version
}

func newConfigWatchEvent(cate WatchEventType, version uint64, val Value) WatchEvent {
	return &watchEvent{cate: cate, version: version, value: val}
}

func NewConfig(keyDelimiter string) Config {
	c := &config{
		keyDelimiter: keyDelimiter,
		watchers:     map[string][]func(event WatchEvent){},
	}
	c.vs.Store(versionValues{version: c.version, values: newValues(keyDelimiter, nil)})
	for i := range c.sourceData {
		c.sourceData[i] = map[string]interface{}{}
	}
	return c
}
