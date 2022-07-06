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
