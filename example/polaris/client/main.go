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
	"time"

	"github.com/imkuqin-zw/yggdrasil"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/polaris/grpc"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/promethues"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/client/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/governor"
)

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		log.Fatal(err)
	}
	go yggdrasil.Run("example.polaris.client")
	client := grpcimpl.NewGreeterClient(grpc.Dial("example.polaris.server"))
	f := func() {
		res, err := client.SayHello(context.TODO(), &helloword.HelloRequest{Name: "fdasf"})
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("call res: %s", res.Message)
		}
	}
	t := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-t.C:
			f()
			t.Reset(time.Second * 2)
		}
	}
}
