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
	"runtime/debug"
)

func Init() {
	handleFunc()
}

func routesHandle(resp http.ResponseWriter, _ *http.Request) {
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
	HandleFunc("/env", envHandle)
	HandleFunc("/configs", configHandle)
	if info, ok := debug.ReadBuildInfo(); ok {
		HandleFunc("/build_info", newBuildInfoHandle(info))
	}
}
