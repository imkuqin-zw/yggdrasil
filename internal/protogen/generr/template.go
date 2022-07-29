package genrpc

import (
	"bytes"
	"strings"
	"text/template"
)

var tpl = `
{{$domain := .Domain}}
{{$codePkg := .CodePackage}}
{{range .Reason}}
var {{.Name}}_code = map[int32]{{$codePkg}}Code{
{{- range $reason, $code := .Codes}}
	{{$reason}}: {{$codePkg}}Code({{$code}}),
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
	Codes map[int32]uint32
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
