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

var serverTpl = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

type {{$svrType}}Server interface {
{{range .Methods -}}
	{{ if .ClientStream -}}
	{{.Name}}({{$svrType}}{{.Name}}Server) error
	{{else if .ServerStream -}}
	{{.Name}}(*{{.Input}}, {{$svrType}}{{.Name}}Server) error
	{{else -}}
	{{.Name}}({{$.Context}}, *{{.Input}}) (*{{.Output}}, error)
	{{ end -}}
{{end -}}
	Unsafe{{$svrType}}Server
}

{{range .Methods}}
{{ if or .ClientStream .ServerStream -}}
type {{$svrType}}{{.Name}}Server interface{
	{{if .ServerStream -}}
	Send(*{{.Output}}) error
	{{end -}}
	{{if .ClientStream -}}
	{{if not .ServerStream -}}
	SendAndClose(*{{.Output}}) error
	{{end -}}
	Recv() (*{{.Input}}, error)
	{{end -}}
	{{$.Types}}ServerStream
}
{{end -}}
{{end}}

type Unsafe{{$svrType}}Server interface {
	mustEmbedUnimplemented{{$svrType}}Server()
}

// Unimplemented{{$svrType}}Server must be embedded to have forward compatible implementations.
type Unimplemented{{$svrType}}Server struct {
}

{{range .Methods -}}
{{ if .ClientStream -}}
func (Unimplemented{{$svrType}}Server) {{.Name}}({{$svrType}}{{.Name}}Server) error {
	return {{$.Errors}}Errorf({{$.Code}}Code_UNAUTHENTICATED, "method {{.Name}} not implemented")
}

{{else if .ServerStream -}}
func (Unimplemented{{$svrType}}Server) {{.Name}}(*{{.Input}}, {{$svrType}}{{.Name}}Server) error{
	return {{$.Errors}}Errorf({{$.Code}}Code_UNAUTHENTICATED, "method {{.Name}} not implemented")
}

{{else -}}
func (Unimplemented{{$svrType}}Server) {{.Name}}({{$.Context}}, *{{.Input}}) (*{{.Output}}, error) {
	return nil, {{$.Errors}}Errorf({{$.Code}}Code_UNAUTHENTICATED, "method {{.Name}} not implemented")
}

{{end -}}
{{end -}}
func (Unimplemented{{$svrType}}Server) mustEmbedUnimplemented{{$svrType}}Server(){}
`
