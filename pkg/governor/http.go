package governor

import (
	"net/http"
)

var (
	// DefaultServeMux ...
	DefaultServeMux = http.NewServeMux()
	routes          = []string{}
)

// HandleFunc ...
func HandleFunc(pattern string, handler http.HandlerFunc) {
	DefaultServeMux.HandleFunc(pattern, handler)
	routes = append(routes, pattern)
}
