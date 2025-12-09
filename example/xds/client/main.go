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

	"github.com/imkuqin-zw/yggdrasil"
	hellowordpb "github.com/imkuqin-zw/yggdrasil/example/protogen/helloword"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	// Import xDS contrib module
	_ "github.com/imkuqin-zw/yggdrasil/contrib/xds"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc"
)

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		logger.Fatal(err)
	}
	yggdrasil.Init("yggdrasil.example.xds.client")
	client := hellowordpb.NewGreeterClient(yggdrasil.NewClient("yggdrasil.example.xds.server"))
	ctx := metadata.WithOutContext(context.Background(), metadata.New(map[string]string{"node": "a"}))
	res, err := client.SayHello(ctx, &hellowordpb.HelloRequest{Name: "fdasf"})
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("call success, resp: %s", res.Message)

	ctx = metadata.WithOutContext(context.Background(), metadata.New(map[string]string{"node": "b"}))
	res, err = client.SayHello(ctx, &hellowordpb.HelloRequest{Name: "fdasf"})
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("call success, resp: %s", res.Message)
}
