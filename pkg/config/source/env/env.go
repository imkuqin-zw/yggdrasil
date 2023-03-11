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

package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/imkuqin-zw/yggdrasil/pkg/config/source"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xarray"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
)

type env struct {
	prefixes         []string
	strippedPrefixes []string
}

func (e *env) Read() (source.SourceData, error) {
	var result = make(map[string]interface{})
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		value := pair[1]
		key := strings.ToLower(pair[0])
		if len(e.prefixes) > 0 || len(e.strippedPrefixes) > 0 {
			notFound := true
			if _, ok := e.matchPrefix(e.prefixes, key); ok {
				notFound = false
			}
			if match, ok := e.matchPrefix(e.strippedPrefixes, key); ok {
				key = strings.TrimPrefix(key, match+"_")
				notFound = false
			}
			if notFound {
				continue
			}
		}
		keys := strings.Split(key, "_")
		xarray.ReverseStringArray(keys)
		tmp := make(map[string]interface{})
		for i, k := range keys {
			if i == 0 {
				if intValue, err := strconv.Atoi(value); err == nil {
					tmp[k] = intValue
				} else if boolValue, err := strconv.ParseBool(value); err == nil {
					tmp[k] = boolValue
				} else {
					tmp[k] = value
				}
				continue
			}
			tmp = map[string]interface{}{k: tmp}
		}
		xmap.MergeStringMap(result, tmp)
	}

	cs := source.NewMapSourceData(source.PriorityEnv, result)
	return cs, nil
}

func (e *env) matchPrefix(pre []string, s string) (string, bool) {
	for _, p := range pre {
		if xstrings.HasPrefix(s, p, "_") {
			return p, true
		}
	}

	return "", false
}

func (e *env) Changeable() bool {
	return false
}

func (e *env) Watch() (<-chan source.SourceData, error) {
	return nil, nil
}

func (e *env) Name() string {
	return "env"
}

func (e *env) Close() error {
	return nil
}

func NewSource(pre, sp []string) source.Source {
	for i, item := range pre {
		pre[i] = strings.ToLower(item)
	}

	for i, item := range sp {
		sp[i] = strings.ToLower(item)
	}
	return &env{prefixes: pre, strippedPrefixes: sp}
}
