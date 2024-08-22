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
	"github.com/imkuqin-zw/yggdrasil/pkg"
)

type Endpoint interface {
	Scheme() string
	Address() string
	Metadata() map[string]string
	Kind() pkg.ServerKind
}

type Server interface {
	RegisterService(sd *ServiceDesc, ss interface{})
	RegisterRestService(sd *RestServiceDesc, ss interface{}, prefix ...string)
	RegisterRestRawHandlers(sd ...*RestRawHandlerDesc)
	Serve(startFlag chan<- struct{}) error
	Stop() error
	Endpoints() []Endpoint
}
