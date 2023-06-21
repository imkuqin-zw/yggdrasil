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

	"google.golang.org/genproto/googleapis/rpc/code"
)

var tpl = `
{{$domain := .Domain}}
{{$codePkg := .CodePackage}}
{{range .Reason}}
var {{.Name}}_code = map[int32]{{$codePkg}}Code{
{{- range $reason, $code := .Codes}}
	{{$reason}}: {{$codePkg}}Code_{{$code}},
{{- end}}
}

func (r {{.Name}}) Reason() string {
	return {{.Name}}_name[int32(r)]
}

func (r {{.Name}}) Domain() string {
	return "{{$domain}}"
}

func (r {{.Name}}) Code() {{$codePkg}}Code {
	return {{.Name}}_code[int32(r)]
}
{{end}}
`

type Reasons struct {
	Domain      string
	CodePackage string
	Reason      []ReasonWrapper
}

type ReasonWrapper struct {
	Name  string
	Codes map[int32]code.Code
}

func (sd *Reasons) execute() string {
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
