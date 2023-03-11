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

package genrpc

import (
	"bytes"
	"strings"
	"text/template"
)

type serviceDesc struct {
	Filename              string
	ServiceType           string
	ServiceName           string
	FullServerName        string
	LowerFirstServiceType string
	Methods               []*methodDesc
	Context               string
	Status                string
	Code                  string
	Client                string
	Server                string
	Interceptor           string
	Md                    string
	Stream                string
}

type methodDesc struct {
	Name         string
	Input        string
	Output       string
	ClientStream bool
	ServerStream bool
}

func (sd *serviceDesc) execute(tpl string) string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("tpl").Parse(strings.TrimSpace(tpl))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, sd); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
