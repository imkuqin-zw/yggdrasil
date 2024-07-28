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

var tpl = commonTpl + clientTpl + serverTpl + descTpl

var commonTpl = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}
{{$lrSvrName := .LowerFirstServiceType}}
{{$ctx := .Context}}
{{$client := .Client}}
{{$metadata := .Md}}
{{$server := .Server}}
{{$interceptor := .Interceptor}}
{{$status := .Status}}
`

var clientTpl = `
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
	{{$.Stream}}ClientStream
}
{{end -}}
{{end}}

type {{$lrSvrName}}Client struct {
	cc {{$client}}Client
}

func New{{$svrType}}Client(cc {{$client}}Client) {{$svrType}}Client {
	return &{{$lrSvrName}}Client{cc}
}

{{range .Methods -}}
{{ if .ClientStream -}}
func (c *{{$lrSvrName}}Client) {{.Name}}(ctx {{$ctx}}) ({{$svrType}}{{.Name}}Client, error) {
	stream, err := c.cc.NewStream(ctx, &{{$svrType}}ServiceDesc.Streams[{{ if .ServerStream -}}0{{else}}1{{end}}], "/{{$.FullServerName}}/{{.Name}}")
	if err != nil {
		return nil, err
	}
	x := &{{$lrSvrName}}{{.Name}}Client{stream}
	return x, nil
}

type {{$lrSvrName}}{{.Name}}Client struct {
	{{$.Stream}}ClientStream
}

func (x *{{$lrSvrName}}{{.Name}}Client) Header() ({{$metadata}}MD, error) {
	v, err := x.ClientStream.Header()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (x *{{$lrSvrName}}{{.Name}}Client) Trailer() {{$metadata}}MD {
	return x.ClientStream.Trailer()
}

func (x *{{$lrSvrName}}{{.Name}}Client) Send(m *{{.Input}}) error {
	return x.ClientStream.SendMsg(m)
}

{{ if .ServerStream -}}
func (x *{{$lrSvrName}}{{.Name}}Client) Recv() (*{{.Output}}, error) {
	m := new({{.Output}})
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
{{end -}}
{{else if .ServerStream -}}
func (c *{{$lrSvrName}}Client) {{.Name}}(ctx {{$ctx}}, in *{{.Input}}) ({{$svrType}}{{.Name}}Client, error) {
	stream, err := c.cc.NewStream(ctx, &{{$svrType}}ServiceDesc.Streams[2], "/{{$.FullServerName}}/{{.Name}}")
	if err != nil {
		return nil, err
	}
	x := &{{$lrSvrName}}{{.Name}}Client{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type {{$lrSvrName}}{{.Name}}Client struct {
	{{$.Stream}}ClientStream
}

func (x *{{$lrSvrName}}{{.Name}}Client) CloseAndRecv() (*{{.Output}}, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new({{.Output}})
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *{{$lrSvrName}}{{.Name}}Client) Recv() (*{{.Output}}, error) {
	m := new({{.Output}})
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *{{$lrSvrName}}{{.Name}}Client) Header() ({{$metadata}}MD, error) {
	v, err := x.ClientStream.Header()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (x *{{$lrSvrName}}{{.Name}}Client) Trailer() {{$metadata}}MD {
	return x.ClientStream.Trailer()
}

func (x *{{$lrSvrName}}{{.Name}}Client) Send(m *{{.Input}}) error {
	return x.ClientStream.SendMsg(m)
}

{{else -}}
func (c *{{$lrSvrName}}Client) {{.Name}}(ctx {{$ctx}}, in *{{.Input}}) (*{{.Output}}, error) {
	out := new({{.Output}})
	err := c.cc.Invoke(ctx, "/{{$.FullServerName}}/{{.Name}}", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

{{end -}}
{{end -}}

{{range .Methods -}}
{{ if .ClientStream -}}
func _{{$svrType}}_{{.Name}}_Handler(srv interface{}, stream {{$.Stream}}ServerStream) error {
	return srv.({{$svrType}}Server).{{.Name}}(&{{$lrSvrName}}{{.Name}}Server{stream})
}

type {{$lrSvrName}}{{.Name}}Server struct {
	{{$.Stream}}ServerStream
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetHeader(md {{$metadata}}MD) error {
	return x.ServerStream.SetHeader(md)
}

func (x *{{$lrSvrName}}{{.Name}}Server) SendHeader(md {{$metadata}}MD) error {
	return x.ServerStream.SendHeader(md)
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetTrailer(md {{$metadata}}MD) {
	x.ServerStream.SetTrailer(md)
}

func (x *{{$lrSvrName}}{{.Name}}Server) Context() {{$ctx}} {
	return x.ServerStream.Context()
}

{{ if .ServerStream -}}
func (x *{{$lrSvrName}}{{.Name}}Server) Send(m *{{.Output}}) error {
	return x.ServerStream.SendMsg(m)
}

{{else -}}
func (x *{{$lrSvrName}}{{.Name}}Server) SendAndClose(m *{{.Output}}) ( error) {
	return x.ServerStream.SendMsg(m)
}

{{end -}}
func (x *{{$lrSvrName}}{{.Name}}Server) Recv() (*{{.Input}}, error) {
	m := new({{.Input}})
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

{{else if .ServerStream -}}
func _{{$svrType}}_{{.Name}}_Handler(srv interface{}, stream {{$.Stream}}ServerStream) error {
	m := new({{.Input}})
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.({{$svrType}}Server).{{.Name}}(m, &{{$lrSvrName}}{{.Name}}Server{stream})
}

type {{$lrSvrName}}{{.Name}}Server struct {
	{{$.Stream}}ServerStream
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetHeader(md {{$metadata}}MD) error {
	return x.ServerStream.SetHeader(md)
}

func (x *{{$lrSvrName}}{{.Name}}Server) SendHeader(md {{$metadata}}MD) error {
	return x.ServerStream.SendHeader(md)
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetTrailer(md {{$metadata}}MD) {
	x.ServerStream.SetTrailer(md)
}

func (x *{{$lrSvrName}}{{.Name}}Server) Context() {{$ctx}} {
	return x.ServerStream.Context()
}

func (x *{{$lrSvrName}}{{.Name}}Server) Send(m *{{.Output}}) error {
	return x.ServerStream.SendMsg(m)
}

{{else -}}
func _{{$svrType}}_{{.Name}}_Handler(srv interface{}, ctx {{$ctx}}, dec func(interface{}) error, unaryInt {{$interceptor}}UnaryServerInterceptor) (interface{}, error) {
	in := new({{.Input}})
	if err := dec(in); err != nil {
		return nil, err
	}
	if unaryInt == nil {
		return srv.({{$svrType}}Server).{{.Name}}(ctx, in)
	}
	info := &{{$interceptor}}UnaryServerInfo{
		Server:     srv,
		FullMethod: "/{{$.FullServerName}}/{{.Name}}",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.({{$svrType}}Server).{{.Name}}(ctx, req.(*{{.Input}}))
	}
	return unaryInt(ctx, in, info, handler)
}

{{end -}}
{{end -}}
`

var serverTpl = `
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
	{{$.Stream}}ServerStream
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
	return {{$status}}Errorf({{$.Code}}Code_UNIMPLEMENTED, "method {{.Name}} not implemented")
}

{{else if .ServerStream -}}
func (Unimplemented{{$svrType}}Server) {{.Name}}(*{{.Input}}, {{$svrType}}{{.Name}}Server) error{
	return {{$status}}Errorf({{$.Code}}Code_UNIMPLEMENTED, "method {{.Name}} not implemented")
}

{{else -}}
func (Unimplemented{{$svrType}}Server) {{.Name}}({{$.Context}}, *{{.Input}}) (*{{.Output}}, error) {
	return nil, {{$status}}Errorf({{$.Code}}Code_UNIMPLEMENTED, "method {{.Name}} not implemented")
}

{{end -}}
{{end -}}
func (Unimplemented{{$svrType}}Server) mustEmbedUnimplemented{{$svrType}}Server(){}
`

var descTpl = `
var {{$svrType}}ServiceDesc = {{$server}}ServiceDesc{
	ServiceName: "{{.FullServerName}}",
	HandlerType: (*{{$svrType}}Server)(nil),
	Methods: []{{$server}}MethodDesc{
		{{range .Methods -}}
		{{if and (not .ClientStream) (not .ServerStream) -}}
		{
			MethodName: "{{.Name}}",
			Handler:    _{{$svrType}}_{{.Name}}_Handler,
		},
		{{end -}}
		{{end -}}
	},
{{if $.NeedStream -}}
	Streams: []{{$.Stream}}StreamDesc{
		{{range .Methods -}}
		{{if or .ClientStream .ServerStream -}}
		{
			StreamName: "{{.Name}}",
			Handler:    _{{$svrType}}_{{.Name}}_Handler,
			{{if .ServerStream -}}
			ServerStreams: true,
			{{end -}}
			{{if .ClientStream -}}
			ClientStreams: true,
			{{end -}}
		},
		{{end -}}
		{{end -}}
	},
{{end -}}
	Metadata: "{{.Filename}}",
}
`
