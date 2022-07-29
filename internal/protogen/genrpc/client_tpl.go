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

var clientTpl = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

type {{$svrType}}Client interface {
{{range .Methods -}}
	{{ if .ClientStream -}}
	{{.Name}}({{$.Context}}) ({{$svrType}}{{.Name}}Client, error)
	{{else if .ServerStream -}}
	{{.Name}}({{$.Context}}, *{{.Input}}) ({{$svrType}}{{.Name}}Client, error)
	{{else -}}
	{{.Name}}({{$.Context}}, *{{.Input}}) (*{{.Output}}, error)
	{{ end -}}
{{end -}}
}

{{range .Methods}}
{{ if or .ClientStream .ServerStream -}}
type {{$svrType}}{{.Name}}Client interface{
	{{if .ClientStream -}}
	Send(*{{.Input}}) error
	{{end -}}
	{{if .ServerStream -}}
	{{if not .ClientStream -}}
	CloseAndRecv() (*{{.Output}}, error)
	{{end -}}
	Recv() (*{{.Output}}, error)
	{{end -}}
	{{$.Types}}ClientStream
}
{{end -}}
{{end}}
`
