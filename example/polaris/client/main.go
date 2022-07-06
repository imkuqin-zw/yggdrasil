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
