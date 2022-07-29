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

import "context"

type RegistryConstructor func() Registry

type Registry interface {
	Register(context.Context, RegistryInstance) error
	Deregister(context.Context, RegistryInstance) error
	Name() string
}

type RegistryInstance interface {
	Region() string
	Zone() string
	Campus() string
	Namespace() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoints() []ServerInfo
}
