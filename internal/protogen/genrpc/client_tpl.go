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
