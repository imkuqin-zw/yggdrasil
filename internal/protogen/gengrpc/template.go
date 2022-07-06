package gengrpc

import (
	"bytes"
	"strings"
	"text/template"
)

var tpl = commonTpl + clientTpl + serverTpl + descTpl

var commonTpl = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}
{{$lrSvrName := .LowerFirstServiceType}}
{{$baseImport := .InterfaceImport}}
{{$ctx := .Context}}
{{$grpc := .Grpc}}
{{$md := .Md}}
{{$grpcMd := .GrpcMd}}
`

var clientTpl = `
type {{$lrSvrName}}Client struct {
	cc {{$grpc}}ClientConnInterface
}

func New{{$svrType}}Client(cc {{$grpc}}ClientConnInterface) {{$baseImport}}{{$svrType}}Client {
	return &{{$lrSvrName}}Client{cc}
}

{{range .Methods -}}
{{ if .ClientStream -}}
func (c *{{$lrSvrName}}Client) {{.Name}}(ctx {{$ctx}}) ({{$baseImport}}{{$svrType}}{{.Name}}Client, error) {
	stream, err := c.cc.NewStream(ctx, &{{$svrType}}ServiceDesc.Streams[{{ if .ServerStream -}}0{{else}}1{{end}}], "{{.FullName}}")
	if err != nil {
		return nil, err
	}
	x := &{{$lrSvrName}}{{.Name}}Client{stream}
	return x, nil
}

type {{$lrSvrName}}{{.Name}}Client struct {
	{{$grpc}}ClientStream
}

func (x *{{$lrSvrName}}{{.Name}}Client) Header() ({{$md}}MD, error) {
	v, err := x.ClientStream.Header()
	if err != nil {
		return nil, err
	}
	return {{$md}}MD(v), nil
}

func (x *{{$lrSvrName}}{{.Name}}Client) Trailer() {{$md}}MD {
	return {{$md}}MD(x.ClientStream.Trailer())
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
func (c *{{$lrSvrName}}Client) {{.Name}}(ctx {{$ctx}}, in *{{.Input}}) ({{$baseImport}}{{$svrType}}{{.Name}}Client, error) {
	stream, err := c.cc.NewStream(ctx, &{{$svrType}}ServiceDesc.Streams[2], "{{.FullName}}")
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
	{{$grpc}}ClientStream
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

func (x *{{$lrSvrName}}{{.Name}}Client) Header() ({{$md}}MD, error) {
	v, err := x.ClientStream.Header()
	if err != nil {
		return nil, err
	}
	return {{$md}}MD(v), nil
}

func (x *{{$lrSvrName}}{{.Name}}Client) Trailer() {{$md}}MD {
	return {{$md}}MD(x.ClientStream.Trailer())
}

func (x *{{$lrSvrName}}{{.Name}}Client) Send(m *{{.Input}}) error {
	return x.ClientStream.SendMsg(m)
}

{{else -}}
func (c *{{$lrSvrName}}Client) {{.Name}}(ctx {{$ctx}}, in *{{.Input}}) (*{{.Output}}, error) {
	out := new({{.Output}})
	err := c.cc.Invoke(ctx, "{{.FullName}}", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

{{end -}}
{{end -}}
`

var serverTpl = `
{{range .Methods -}}
{{ if .ClientStream -}}
func _{{$svrType}}_{{.Name}}_Handler(srv interface{}, stream {{$grpc}}ServerStream) error {
	return srv.({{$baseImport}}{{$svrType}}Server).{{.Name}}(&{{$lrSvrName}}{{.Name}}Server{stream})
}

type {{$lrSvrName}}{{.Name}}Server struct {
	{{$grpc}}ServerStream
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetHeader(md md.MD) error {
	return x.ServerStream.SetHeader({{$grpcMd}}MD(md))
}

func (x *{{$lrSvrName}}{{.Name}}Server) SendHeader(md md.MD) error {
	return x.ServerStream.SendHeader({{$grpcMd}}MD(md))
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetTrailer(md md.MD) {
	x.ServerStream.SetTrailer({{$grpcMd}}MD(md))
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
func _{{$svrType}}_{{.Name}}_Handler(srv interface{}, stream {{$grpc}}ServerStream) error {
	m := new({{.Input}})
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.({{$baseImport}}{{$svrType}}Server).{{.Name}}(m, &{{$lrSvrName}}{{.Name}}Server{stream})
}

type {{$lrSvrName}}{{.Name}}Server struct {
	{{$grpc}}ServerStream
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetHeader(md md.MD) error {
	return x.ServerStream.SetHeader({{$grpcMd}}MD(md))
}

func (x *{{$lrSvrName}}{{.Name}}Server) SendHeader(md md.MD) error {
	return x.ServerStream.SendHeader({{$grpcMd}}MD(md))
}

func (x *{{$lrSvrName}}{{.Name}}Server) SetTrailer(md md.MD) {
	x.ServerStream.SetTrailer({{$grpcMd}}MD(md))
}

func (x *{{$lrSvrName}}{{.Name}}Server) Context() {{$ctx}} {
	return x.ServerStream.Context()
}

func (x *{{$lrSvrName}}{{.Name}}Server) Send(m *{{.Output}}) error {
	return x.ServerStream.SendMsg(m)
}

{{else -}}
func _{{$svrType}}_{{.Name}}_Handler(srv interface{}, ctx {{$ctx}}, dec func(interface{}) error, interceptor {{$grpc}}UnaryServerInterceptor) (interface{}, error) {
	in := new({{.Input}})
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.({{$baseImport}}{{$svrType}}Server).{{.Name}}(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "{{.FullName}}",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.({{$baseImport}}{{$svrType}}Server).{{.Name}}(ctx, req.(*{{.Input}}))
	}
	return interceptor(ctx, in, info, handler)
}

{{end -}}
{{end -}}
`

var descTpl = `
var {{$svrType}}ServiceDesc = {{$grpc}}ServiceDesc{
	ServiceName: "{{.FullServerName}}",
	HandlerType: (*{{$baseImport}}{{$svrType}}Server)(nil),
	Methods: []{{$grpc}}MethodDesc{
		{{range .Methods -}}
		{{if and (not .ClientStream) (not .ServerStream) -}}
		{
			MethodName: "{{.Name}}",
			Handler:    _{{$svrType}}_{{.Name}}_Handler,
		},
		{{end -}}
		{{end -}}
	},
	Streams: []{{$grpc}}StreamDesc{
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
	Metadata: "{{.Filename}}",
}
`

type serviceDesc struct {
	Filename              string
	ServiceType           string
	LowerFirstServiceType string
	InterfaceImport       string
	ServiceName           string
	FullServerName        string
	Status                string
	Grpc                  string
	Md                    string
	GrpcMd                string
	Context               string
	Code                  string
	Methods               []*methodDesc
}

type methodDesc struct {
	Name         string
	FullName     string
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
