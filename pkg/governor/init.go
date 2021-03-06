// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
