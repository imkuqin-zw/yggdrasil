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
	"errors"
	"net/http"

	"github.com/imkuqin-zw/yggdrasil"
	librarypb "github.com/imkuqin-zw/yggdrasil/example/protogen/library"
	librarypb2 "github.com/imkuqin-zw/yggdrasil/example/protogen/library/v1"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/interceptor/logging"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
	"github.com/imkuqin-zw/yggdrasil/pkg/status"
)

type LibraryImpl struct {
	librarypb2.UnimplementedLibraryServiceServer
}

func (s *LibraryImpl) CreateShelf(ctx context.Context, request *librarypb2.CreateShelfRequest) (*librarypb2.Shelf, error) {
	_ = metadata.SetTrailer(ctx, metadata.Pairs("trailer", "test"))
	_ = metadata.SetHeader(ctx, metadata.Pairs("header", "test"))
	return &librarypb2.Shelf{Name: "test", Theme: "test"}, nil
}

func (s *LibraryImpl) GetShelf(ctx context.Context, request *librarypb2.GetShelfRequest) (*librarypb2.Shelf, error) {
	_ = metadata.SetTrailer(ctx, metadata.Pairs("trailer", "test"))
	_ = metadata.SetHeader(ctx, metadata.Pairs("header", "test"))
	return &librarypb2.Shelf{Name: request.Name, Theme: "test"}, nil
}

func (s *LibraryImpl) MoveBook(ctx context.Context, request *librarypb2.MoveBookRequest) (*librarypb2.Book, error) {
	return nil, status.FromReason(errors.New("test reason"), librarypb.Reason_BOOK_NOT_FOUND, nil)
}

func (s *LibraryImpl) GetBook(ctx context.Context, request *librarypb2.GetBookRequest) (*librarypb2.Book, error) {
	return &librarypb2.Book{Name: request.Name}, nil
}

func WebHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello web"))
}

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		logger.FatalField("fault to load config file", logger.Err(err))
	}
	yggdrasil.Init("github.com.imkuqin_zw.yggdrasil.example.sample")
	ss := &LibraryImpl{}

	if err := yggdrasil.Serve(
		yggdrasil.WithServiceDesc(&librarypb2.LibraryServiceServiceDesc, ss),
		yggdrasil.WithRestServiceDesc(&librarypb2.LibraryServiceRestServiceDesc, ss),
		yggdrasil.WithRestRawHandleDesc(&server.RestRawHandlerDesc{
			Method:  http.MethodGet,
			Path:    "/web",
			Handler: WebHandler,
		}),
	); err != nil {
		logger.FatalField("the application was ended forcefully ", logger.Err(err))
		logger.Fatal(err)
	}
}
