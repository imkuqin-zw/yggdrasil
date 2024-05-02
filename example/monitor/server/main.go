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
	"encoding/json"
	"fmt"
	"time"

	"github.com/imkuqin-zw/yggdrasil"
	xgorm "github.com/imkuqin-zw/yggdrasil/contrib/gorm"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/gorm/driver/mysql"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/gorm/plugin/metrics"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/gorm/plugin/trace"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/otelexporters/otlpgrpc"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/otelexporters/prometheus"
	xredis "github.com/imkuqin-zw/yggdrasil/contrib/redis"
	"github.com/imkuqin-zw/yggdrasil/example/protogen/helloword"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/interceptor/logging"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/stats/otel"
	"gorm.io/gorm"
)

type Greeter struct {
	Pk  int `gorm:"primaryKey"`
	Val int `gorm:"column:val"`
}

type GreeterImpl struct {
	db    *gorm.DB
	cache xredis.Redis
	helloword.UnimplementedGreeterServer
}

func (g GreeterImpl) SayHello(ctx context.Context, request *helloword.HelloRequest) (*helloword.HelloReply, error) {
	m := &Greeter{Pk: 1, Val: 1234}
	session := g.db.WithContext(ctx)
	if err := session.Save(m).Error; err != nil {
		return nil, err
	}
	if err := session.Where("pk = ?", 1).First(m).Error; err != nil {
		return nil, err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err = g.cache.Set(ctx, "greeter", data, time.Second*5).Err(); err != nil {
		return nil, err
	}
	data, err = g.cache.Get(ctx, "greeter").Bytes()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	return &helloword.HelloReply{Message: fmt.Sprintf("%s, %d", request.Name, m.Val)}, nil
}

func main() {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		logger.FatalField("fault to load config file", logger.Err(err))
	}
	yggdrasil.Init("github.com.imkuqin_zw.yggdrasil.example.sample")
	db := xgorm.NewDB("default")
	cache := xredis.NewRedis("default")
	service := GreeterImpl{
		db:    db,
		cache: cache,
	}
	if err := yggdrasil.Serve(yggdrasil.WithServiceDesc(&helloword.GreeterServiceDesc, service)); err != nil {
		logger.FatalField("the application was ended forcefully ", logger.Err(err))
		logger.Fatal(err)
	}
}
