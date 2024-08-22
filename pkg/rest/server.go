package rest

import (
	"net/http"
)

// A HandlerFunc handles a specific pair of path pattern and HTTP method.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) (interface{}, error)

type ServerInfo interface {
	GetAddress() string
	GetAttributes() map[string]string
}

type Server interface {
	RpcHandle(method, path string, f HandlerFunc)
	RawHandle(method, path string, h http.HandlerFunc)
	Start() error
	Serve() error
	Stop() error
	Info() ServerInfo
}
