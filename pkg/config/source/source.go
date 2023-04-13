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
	"io"

	"github.com/mitchellh/mapstructure"
)

type Priority uint8

const (
	PriorityFile Priority = iota
	PriorityEnv
	PriorityFlag
	PriorityCli
	PriorityRemote
	PriorityMemory
	PriorityMax
)

type SourceData interface {
	Priority() Priority
	Data() []byte
	Unmarshal(v interface{}) error
}

// Source is the source from which conf is loaded
type Source interface {
	Name() string
	Read() (SourceData, error)
	Changeable() bool
	Watch() (<-chan SourceData, error)
	io.Closer
}

type bytesSourceData struct {
	priority  Priority
	data      []byte
	unmarshal func([]byte, interface{}) error
}

func NewBytesSourceData(priority Priority, data []byte,
	unmarshal func([]byte, interface{}) error) SourceData {
	return &bytesSourceData{priority: priority, data: data, unmarshal: unmarshal}
}

func (c *bytesSourceData) Priority() Priority {
	return c.priority
}

func (c *bytesSourceData) Data() []byte {
	return c.data
}

func (c *bytesSourceData) Unmarshal(v interface{}) error {
	return c.unmarshal(c.data, v)
}

type mapSourceData struct {
	priority  Priority
	data      map[string]interface{}
	unmarshal func([]byte, interface{}) error
}

func NewMapSourceData(priority Priority, data map[string]interface{}) SourceData {
	return &mapSourceData{priority: priority, data: data}
}

func (c *mapSourceData) Priority() Priority {
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
		TagName:    "yaml",
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	return decoder.Decode(c.data)
}
