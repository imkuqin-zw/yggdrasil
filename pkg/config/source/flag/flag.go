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

package flag

import (
	flag2 "flag"
	"os"
	"strings"

	"github.com/imkuqin-zw/yggdrasil/pkg/config/source"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
)

type flag struct {
	fs *flag2.FlagSet
}

func (fs *flag) Read() (source.SourceData, error) {
	if !fs.fs.Parsed() {
		_ = fs.fs.Parse(os.Args[1:])
		for len(fs.fs.Args()) != 0 {
			_ = fs.fs.Parse(fs.fs.Args()[1:])
		}
	}
	var result = make(map[string]interface{})
	visitFn := func(f *flag2.Flag) {
		n := strings.ToLower(f.Name)
		keys := strings.FieldsFunc(n, split)
		reverse(keys)

		tmp := make(map[string]interface{})
		for i, k := range keys {
			if i == 0 {
				if v, ok := f.Value.(flag2.Getter); ok {
					tmp[k] = v.Get()
				} else {
					tmp[k] = f.Value
				}
				continue
			}

			tmp = map[string]interface{}{k: tmp}
		}
		xmap.MergeStringMap(result, tmp)
		return
	}
	fs.fs.VisitAll(visitFn)
	cs := source.NewMapSourceData(source.PriorityFlag, result)
	return cs, nil
}

func split(r rune) bool {
	return r == '-' || r == '_'
}

func reverse(ss []string) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}

func (fs *flag) Changeable() bool {
	return false
}

func (fs *flag) Watch() (<-chan source.SourceData, error) {
	return nil, nil
}

func (fs *flag) Name() string {
	return "env"
}

func (fs *flag) Close() error {
	return nil
}

func NewSource(fs ...*flag2.FlagSet) source.Source {
	if len(fs) == 0 || fs[0] == nil {
		return &flag{fs: flag2.CommandLine}
	}
	return &flag{fs: fs[0]}
}
