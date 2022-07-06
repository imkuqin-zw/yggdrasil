package genrpc

import (
	"bytes"
	"strings"
	"text/template"
)

type serviceDesc struct {
	ServiceType string
	ServiceName string
	Methods     []*methodDesc
	Types       string
	Context     string
	Status      string
	Errors      string
	Code        string
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
