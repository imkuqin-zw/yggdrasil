package governor

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime/debug"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
)

func Init() {
	handleFunc()
}

func routesHandle(resp http.ResponseWriter, req *http.Request) {
	_ = json.NewEncoder(resp).Encode(routes)
}

func handleFunc() {
	// 获取全部治理路由
	DefaultServeMux.HandleFunc("/", routesHandle)
	HandleFunc("/routes", routesHandle)
	HandleFunc("/debug/pprof/", pprof.Index)
	HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	HandleFunc("/debug/pprof/profile", pprof.Profile)
	HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	HandleFunc("/debug/pprof/trace", pprof.Trace)
	HandleFunc("/configs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		if r.URL.Query().Get("pretty") == "true" {
			encoder.SetIndent("", "    ")
		}
		_ = encoder.Encode(json.RawMessage(config.Bytes()))
	})
	HandleFunc("/debug/env", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		if r.URL.Query().Get("pretty") == "true" {
			encoder.SetIndent("", "    ")
		}
		_ = encoder.Encode(os.Environ())
	})
	if info, ok := debug.ReadBuildInfo(); ok {
		HandleFunc("/mod", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			encoder := json.NewEncoder(w)
			if r.URL.Query().Get("pretty") == "true" {
				encoder.SetIndent("", "    ")
			}
			_ = encoder.Encode(info)
		})
	}
	HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		if r.URL.Query().Get("pretty") == "true" {
			encoder.SetIndent("", "    ")
		}
		_ = encoder.Encode(services)
	})
}
