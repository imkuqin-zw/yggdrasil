package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Builder func() func(http.Handler) http.Handler

var (
	builder = map[string]Builder{}
)

func RegisterBuilder(name string, f Builder) {
	builder[name] = f
}

func GetMiddlewares(names ...string) chi.Middlewares {
	var handlers = make(chi.Middlewares, 0, len(names)+1)
	handlers = append(handlers, newMarshalerMiddleware())
	for _, item := range names {
		if f, ok := builder[item]; ok {
			handlers = append(handlers, f())
		}
	}
	return handlers
}
