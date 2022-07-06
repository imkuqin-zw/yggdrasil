package genrpc

import (
	"bytes"
	"strings"
	"text/template"
)

var tpl = `
{{$domain := .Domain}}
{{range .Reason}}
func (r {{.Name}}) Reason() string {
	return {{.Name}}_name[int32(r)]
}

func (r {{.Name}}) Domain() string {
	return "{{$domain}}"
}
{{end}}
`

type Reasons struct {
	Domain string
	Reason []ReasonWrapper
}

type ReasonWrapper struct {
	Name string
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
