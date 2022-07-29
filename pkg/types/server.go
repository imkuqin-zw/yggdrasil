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

type ServerKind string

const (
	ServerKindRpc      ServerKind = "rpc"
	ServerKindJob      ServerKind = "job"
	ServerKindTask     ServerKind = "task"
	ServerKindGovernor ServerKind = "governor"
)

type ServerInfo interface {
	Scheme() string
	Host() string
	Kind() ServerKind
	Endpoint() string
	Metadata() map[string]string
}

type ServerConstructor func() Server

type Server interface {
	Serve() error
	Stop() error
	Info() ServerInfo
}
