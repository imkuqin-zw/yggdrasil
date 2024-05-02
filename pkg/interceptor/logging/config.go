package logging

import (
	"fmt"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/pkg/errors"
)

type Config struct {
	SlowThreshold  time.Duration `default:"1s"`
	PrintReqAndRes bool
}

func initCfg() {
	cfg := Config{}
	if err := config.Get(fmt.Sprintf(config.KeyInterceptorCfg, name)).Scan(&cfg); err != nil {
		logger.ErrorField("fault to load logger config", logger.Err(errors.WithStack(err)))
	}
}
