// Copyright 2022 The imkuqin-zw Authors.
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

package main

import (
	"net/http"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/interceptor/logging"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
)

func WebSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello web"))
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/web", http.StatusMovedPermanently)
}

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		logger.FatalField("fault to load config file", logger.Err(err))
	}
	yggdrasil.Init("github.com.imkuqin_zw.yggdrasil.example.web")

	if err := yggdrasil.Serve(
		yggdrasil.WithRestRawHandleDesc(
			&server.RestRawHandlerDesc{
				Method:  http.MethodGet,
				Path:    "/web",
				Handler: WebSuccessHandler,
			},
			&server.RestRawHandlerDesc{
				Method:  http.MethodGet,
				Path:    "/redirect",
				Handler: RedirectHandler,
			},
		),
	); err != nil {
		logger.FatalField("the application was ended forcefully ", logger.Err(err))
		logger.Fatal(err)
	}
}
