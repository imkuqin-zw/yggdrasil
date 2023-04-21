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

package config

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"github.com/mitchellh/mapstructure"
)

type values struct {
	keyDelimiter string
	val          map[string]interface{}
}

func newValues(keyDelimiter string, val map[string]interface{}) *values {
	if val == nil {
		val = map[string]interface{}{}
	}
	v, _ := xmap.CloneMap(val)
	return &values{keyDelimiter: keyDelimiter, val: v}
}

func (vs *values) get(key string) interface{} {
	dd, ok := vs.val[key]
	if ok {
		return dd
	}
	return xmap.DeepSearchInMap(vs.val, genPath(key, vs.keyDelimiter)...)
}

func (vs *values) Get(key string) Value {
	if key == "" {
		return &value{val: vs.val}
	}
	return newValue(vs.get(key))
}

func (vs *values) GetMulti(keys ...string) Value {
	result := map[string]interface{}{}
	for _, key := range keys {
		if data, ok := vs.get(key).(map[string]interface{}); ok {
			xmap.MergeStringMap(result, data)
		}
	}
	return newValue(result)
}

func (vs *values) Del(key string) error {
	paths := genPath(key, vs.keyDelimiter)
	tmp := vs.val
	var ok bool
	for _, path := range paths[:len(paths)-1] {
		tmp, ok = tmp[path].(map[string]interface{})
		if !ok {
			return nil
		}
	}
	delete(tmp, key)
	return nil
}

func (vs *values) Set(key string, val interface{}) error {
	paths := genPath(key, vs.keyDelimiter)
	tmp := vs.val
	var ok bool
	for _, path := range paths[:len(paths)-1] {
		tmp, ok = tmp[path].(map[string]interface{})
		if !ok {
			return nil
		}
	}
	tmp[key] = val
	return nil
}

func (vs *values) SetMulti(keys []string, values []interface{}) error {
	if len(keys) != len(values) {
		return errors.New("the quantity of key and value does not match")
	}
	for i := 0; i < len(keys); i++ {
		if err := vs.Set(keys[i], values[i]); err != nil {
			return err
		}
	}
	return nil
}

func (vs *values) Map() map[string]interface{} {
	if vs.val == nil {
		return nil
	}
	val, _ := xmap.CloneMap(vs.val)
	return val
}

func (vs *values) Scan(v interface{}) error {
	c := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     v,
		TagName:    "yaml",
	}
	decoder, err := mapstructure.NewDecoder(&c)
	if err != nil {
		return err
	}
	if err := decoder.Decode(vs.val); err != nil {
		return err
	}
	if reflect.TypeOf(v).Kind() != reflect.Ptr || reflect.ValueOf(v).Elem().Kind() != reflect.Struct {
		return nil
	}
	return defaults.Set(v)
}

func (vs *values) Bytes() []byte {
	if vs.val != nil {
		data, _ := json.Marshal(vs.val)
		return data
	}
	return []byte{}
}

func (vs *values) deepSearchInMap(val map[string]interface{}, key, delimiter string) interface{} {
	if v, ok := val[key]; ok {
		return v
	}
	keys := strings.SplitN(key, delimiter, 2)
	tmp, ok := val[keys[0]].(map[string]interface{})
	if !ok {
		return nil
	}
	return vs.deepSearchInMap(tmp, keys[1], delimiter)
}
