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

package source

import (
	"encoding/json"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/mitchellh/mapstructure"
)

type bytesSourceData struct {
	priority  types.ConfigPriority
	data      []byte
	unmarshal func([]byte, interface{}) error
}

func NewBytesSourceData(priority types.ConfigPriority, data []byte,
	unmarshal func([]byte, interface{}) error) types.ConfigSourceData {
	return &bytesSourceData{priority: priority, data: data, unmarshal: unmarshal}
}

func (c *bytesSourceData) Priority() types.ConfigPriority {
	return c.priority
}

func (c *bytesSourceData) Data() []byte {
	return c.data
}

func (c *bytesSourceData) Unmarshal(v interface{}) error {
	return c.unmarshal(c.data, v)
}

type mapSourceData struct {
	priority  types.ConfigPriority
	data      map[string]interface{}
	unmarshal func([]byte, interface{}) error
}

func NewMapSourceData(priority types.ConfigPriority, data map[string]interface{}) types.ConfigSourceData {
	return &mapSourceData{priority: priority, data: data}
}

func (c *mapSourceData) Priority() types.ConfigPriority {
	return c.priority
}

func (c *mapSourceData) Data() []byte {
	data, _ := json.Marshal(c.data)
	return data
}

func (c *mapSourceData) Unmarshal(v interface{}) error {
	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     v,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	return decoder.Decode(c.data)
}
