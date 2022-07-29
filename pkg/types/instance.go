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

package types

type InstanceInfo interface {
	// 命名空间
	Namespace() string
	// 服务名
	Name() string
	// 版本号
	Version() string
	// 地区
	Region() string
	// 地域
	Zone() string
	// 园区
	Campus() string
	// 元信息
	Metadata() map[string]string
	// 添加元信息（忽略已存在的KEY）
	AddMetadata(key, val string) bool
}
