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
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

var cfg = NewConfig(".")

func init() {
	defers.Register(func() error {
		return cfg.Close()
	})
}

func Get(key string) types.ConfigValue {
	return cfg.Get(key)
}

func Set(key string, val interface{}) error {
	return cfg.Set(key, val)
}

func Bytes() []byte {
	return cfg.Bytes()
}

func GetBool(key string, def ...bool) bool {
	return cfg.Get(key).Bool(def...)
}

func GetInt(key string, def ...int) int {
	return cfg.Get(key).Int(def...)
}

func GetInt64(key string, def ...int64) int64 {
	return cfg.Get(key).Int64(def...)
}

func GetString(key string, def ...string) string {
	return cfg.Get(key).String(def...)
}

func GetBytes(key string, def ...[]byte) []byte {
	return cfg.Get(key).Bytes(def...)
}

func GetStringSlice(key string, def ...[]string) []string {
	return cfg.Get(key).StringSlice(def...)
}

func GetStringMap(key string, def ...map[string]string) map[string]string {
	return cfg.Get(key).StringMap(def...)
}

func GetMap(key string, def ...map[string]interface{}) map[string]interface{} {
	return cfg.Get(key).Map(def...)
}

func GetFloat64(key string, def ...float64) float64 {
	return cfg.Get(key).Float64(def...)
}

func GetDuration(key string, def ...time.Duration) time.Duration {
	return cfg.Get(key).Duration(def...)
}

func Scan(key string, val interface{}) error {
	return cfg.Get(key).Scan(val)
}

func LoadSource(sources ...types.ConfigSource) error {
	return cfg.LoadSource(sources...)
}

func AddWatcher(key string, f func(types.ConfigWatchEvent)) error {
	return cfg.AddWatcher(key, f)
}

func DelWatcher(key string, f func(types.ConfigWatchEvent)) error {
	return cfg.DelWatcher(key, f)
}
