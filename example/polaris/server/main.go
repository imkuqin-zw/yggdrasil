package main

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"

	_ "github.com/imkuqin-zw/yggdrasil/contrib/polaris"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/promethues"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
)

type GreeterImpl struct {
	helloword.UnimplementedGreeterServer
}

func (g GreeterImpl) SayHello(ctx context.Context, request *helloword.HelloRequest) (*helloword.HelloReply, error) {
	return &helloword.HelloReply{Message: request.Name}, nil
}

func (g GreeterImpl) SayHelloStream(server helloword.GreeterSayHelloStreamServer) error {
	panic("implement me")
}

func (g GreeterImpl) SayHelloClientStream(server helloword.GreeterSayHelloClientStreamServer) error {
	panic("implement me")
}

func (g GreeterImpl) SayHelloServerStream(request *helloword.HelloRequest, server helloword.GreeterSayHelloServerStreamServer) error {
	panic("implement me")
}

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		log.Fatal(err)
	}
	grpc.RegisterService(&grpcimpl.GreeterServiceDesc, GreeterImpl{})
	if err := yggdrasil.Run("example.polaris.server"); err != nil {
		log.Fatal(err)
	}
}
