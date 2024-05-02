package xredis

import (
	"context"
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

type logging struct{}

func (l *logging) Printf(ctx context.Context, format string, v ...interface{}) {
	logger.ErrorField(fmt.Sprintf(format, v...), logger.Context(ctx))
}
