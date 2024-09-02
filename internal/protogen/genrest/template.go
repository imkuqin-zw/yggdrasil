package genrest

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

var restTemplate = `
{{range $method := .Methods }}
func local_handler_{{$.ServiceType}}_{{ .Name }}_{{.Num}}(w {{$.HttpPkg}}ResponseWriter, r *{{$.HttpPkg}}Request, server interface{}, unaryInt {{$.InterceptorPkg}}UnaryServerInterceptor) (interface{}, error) {
		protoReq := &{{$method.Request}}{}
		{{if $method.HasBody }}
			inbound := {{$.MarshalerPkg}}InboundFromContext(r.Context())
			if err := inbound.NewDecoder(r.Body).Decode(protoReq{{$method.Body}}); err != nil && err != {{$.IoPkg}}EOF {
				return nil, {{$.StatusPkg}}New({{$.CodePkg}}Code_INVALID_ARGUMENT, err)
			}
		{{else -}}
			if err := {{$.RestPkg}}PopulateQueryParameters(protoReq, r.URL.Query()); err != nil {
				return nil,  {{$.StatusPkg}}New({{$.CodePkg}}Code_INVALID_ARGUMENT, err)
			}
		{{end -}}

		{{- range  $key, $value := .PathVars}}
			if val := {{parsePathValues $value }}; len(val) == 0 {
				return nil, {{$.StatusPkg}}Errorf({{$.CodePkg}}Code_INVALID_ARGUMENT, "not found {{$key}}")
			} else if err := {{$.RestPkg}}PopulateFieldFromPath(protoReq, {{$key | printf "%q"}}, val); err != nil {
				return nil, {{$.StatusPkg}}New({{$.CodePkg}}Code_INVALID_ARGUMENT, err)
			}
		{{- end}}
	
		if unaryInt == nil {
			return  server.({{$.ServiceType}}Server).{{$method.Name}}(r.Context(), protoReq)
		}

		info := &interceptor.UnaryServerInfo{
			Server:     server,
			FullMethod: "{{$.ServiceName}}/{{ .Name }}",
		}
		handler := func(ctx {{$.CtxPkg}}Context, req interface{}) (interface{}, error) {
			return server.({{$.ServiceType}}Server).{{$method.Name}}(ctx, req.(*{{$method.Request}}))
		}
		return unaryInt(r.Context(), protoReq, info, handler)
}
{{end -}}

var {{$.ServiceType}}RestServiceDesc = {{$.SvrPkg}}RestServiceDesc{
	HandlerType: (*{{$.ServiceType}}Server)(nil),
	Methods: []{{$.SvrPkg}}RestMethodDesc{
		{{range $method := .Methods -}}
		{
			Method: "{{$method.Method}}",
			Path: "{{$method.Path}}",
			Handler:    local_handler_{{$.ServiceType}}_{{ .Name }}_{{.Num}},
		},
		{{end -}}
	},
}
`

type serviceDesc struct {
	HttpPkg        string
	ChiPkg         string
	MarshalerPkg   string
	StatusPkg      string
	RestPkg        string
	SvrPkg         string
	CodePkg        string
	InterceptorPkg string
	CtxPkg         string
	IoPkg          string

	ServiceType string
	ServiceName string
	Methods     []*methodDesc
}

type methodDesc struct {
	Name    string
	Num     int
	Method  string
	Request string

	PathVars map[string]string
	Path     string

	Body    string
	HasBody bool
}

func (s *serviceDesc) execute() string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http-rest").
		Funcs(template.FuncMap{"parsePathValues": s.parsePathValues}).
		Parse(strings.TrimSpace(restTemplate))
	if err != nil {
		panic(err)
	}
	if err = tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return string(buf.Bytes())
}

func (s *serviceDesc) parsePathValues(path string) string {
	subPattern0 := regexp.MustCompile(`(?i)^{params[0-9]+}$`)
	if subPattern0.MatchString(path) {
		path = fmt.Sprintf(`%sURLParam(r, "%s")`, s.ChiPkg, strings.TrimRight(strings.TrimLeft(path, "{"), "}"))
		return path
	}
	path = subPattern0.ReplaceAllStringFunc(path, func(subStr string) string {
		params := pathPattern.FindStringSubmatch(subStr)
		return fmt.Sprintf(`%sURLParam(r, "%s")+"/`, s.ChiPkg, params[1])
	})
	subPattern1 := regexp.MustCompile(`(?i)/{(params[0-9]+)}/`)
	path = subPattern1.ReplaceAllStringFunc(path, func(subStr string) string {
		params := pathPattern.FindStringSubmatch(subStr)
		return fmt.Sprintf(`/"+%sURLParam(r, "%s")+"/`, s.ChiPkg, params[1])
	})
	subPattern2 := regexp.MustCompile(`(?i)/{(params[0-9]+)}`)
	path = subPattern2.ReplaceAllStringFunc(path, func(subStr string) string {
		params := pathPattern.FindStringSubmatch(subStr)
		return fmt.Sprintf(`/"+%sURLParam(r, "%s")+"`, s.ChiPkg, params[1])
	})
	path = fmt.Sprintf(`"%s"`, path)
	path = strings.TrimRight(path, `+""`)
	return path
}
