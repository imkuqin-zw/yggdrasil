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
	"flag"
	"os"
	"os/signal"

	"github.com/imkuqin-zw/yggdrasil/internal/prohethues_polaris_sd"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	flag2 "github.com/imkuqin-zw/yggdrasil/pkg/config/source/flag"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
)

var (
	_ = flag.String("conf", "./config.yaml", "config path")
)

func init() {
	if err := config.LoadSource(flag2.NewSource()); err != nil {
		log.Fatalf("fault to load flag source, err: %+v", err)
	}
	cfgFile := config.GetString("conf", "./config.yaml")
	if err := config.LoadSource(file.NewSource(cfgFile, false)); err != nil {
		log.Fatalf("fault to load file source, filepath: %s, err: %+v", cfgFile, err)
	}
}

func main() {
	go shutdown()
	prohethues_polaris_sd.Run()
}

func shutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	log.Info("shutdown")
	os.Exit(0)
}
