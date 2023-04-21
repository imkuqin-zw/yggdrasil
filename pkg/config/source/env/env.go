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
	parseArray       bool
	arraySep         string
	delimiter        string
}

func (e *env) parseValue(value string) interface{} {
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	} else if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	} else if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	} else {
		return value
	}
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
				key = strings.TrimPrefix(key, match+e.delimiter)
				notFound = false
			}
			if notFound {
				continue
			}
		}
		keys := strings.Split(key, e.delimiter)
		xarray.ReverseStringArray(keys)
		tmp := make(map[string]interface{})
		for i, k := range keys {
			if i == 0 {
				if e.parseArray {
					values := strings.Split(value, e.arraySep)
					if len(values) > 1 {
						tmpVal := make([]interface{}, len(values))
						for j, item := range values {
							tmpVal[j] = e.parseValue(item)
						}
						tmp[k] = tmpVal
						continue
					}
				}
				tmp[k] = e.parseValue(value)
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
		if xstrings.HasPrefix(s, p, e.delimiter) {
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

func NewSource(pre, sp []string, opts ...Option) source.Source {
	for i, item := range pre {
		pre[i] = strings.ToLower(item)
	}

	for i, item := range sp {
		sp[i] = strings.ToLower(item)
	}
	e := &env{prefixes: pre, strippedPrefixes: sp, delimiter: "_"}
	for _, opt := range opts {
		opt(e)
	}
	return e
}
