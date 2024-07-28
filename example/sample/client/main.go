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
	"context"
	"fmt"

	"github.com/imkuqin-zw/yggdrasil"
	librarypb "github.com/imkuqin-zw/yggdrasil/example/protogen/library/v1"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/interceptor/logging"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
)

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		logger.Fatal(err)
	}
	yggdrasil.Init("github.com.imkuqin_zw.yggdrasil.example.sample.client")
	client := librarypb.NewLibraryServiceClient(yggdrasil.NewClient("github.com.imkuqin_zw.yggdrasil.example.sample"))
	ctx := metadata.WithStreamContext(context.TODO())
	_, err := client.GetShelf(ctx, &librarypb.GetShelfRequest{Name: "fdasf"})
	if err != nil {
		logger.Fatal(err)
	}
	if trailer, ok := metadata.FromTrailerCtx(ctx); ok {
		fmt.Println(trailer)
	}
	if header, ok := metadata.FromHeaderCtx(ctx); ok {
		fmt.Println(header)
	}
	_, err = client.MoveBook(context.TODO(), &librarypb.MoveBookRequest{Name: "fdasf"})
	if err != nil {
		fmt.Println(status.FromError(err).Reason().Reason)
	}
	logger.Info("call success")
}
