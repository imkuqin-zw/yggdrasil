package main

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/client/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
)

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		log.Fatal(err)
	}
	_ = yggdrasil.Run("client")
	client := grpcimpl.NewGreeterClient(grpc.Dial("sample"))
	res, err := client.SayHello(context.TODO(), &helloword.HelloRequest{Name: "fdasf"})
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("call success, res: %s", res.Message)
}
