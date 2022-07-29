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

package server

import (
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

type info struct {
	scheme   string
	kind     types.ServerKind
	host     string
	metadata map[string]string
}

func (si *info) Endpoint() string {
	return fmt.Sprintf("%s://%s", si.scheme, si.host)
}

func (si *info) Host() string {
	return si.host
}

func (si *info) Metadata() map[string]string {
	return si.metadata
}

func (si *info) Kind() types.ServerKind {
	return si.kind
}

func (si *info) Scheme() string {
	return si.scheme
}

func NewInfo(scheme string, kind types.ServerKind, host string, metadata map[string]string) types.ServerInfo {
	return &info{scheme: scheme, kind: kind, host: host, metadata: metadata}
}
