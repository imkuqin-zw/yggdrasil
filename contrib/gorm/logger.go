package xgorm

import (
	"context"
	"fmt"
	"strings"
	"time"

	lg "github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

type logger struct {
	slowThreshold time.Duration
}

func (l *logger) Info(ctx context.Context, s string, i ...interface{}) {
	lg.InfoField(fmt.Sprintf(strings.TrimRight(s, "\n"), i...), lg.Context(ctx))
}

func (l *logger) Warn(ctx context.Context, s string, i ...interface{}) {
	lg.WarnField(fmt.Sprintf(strings.TrimRight(s, "\n"), i...), lg.Context(ctx))
}

func (l *logger) Error(ctx context.Context, s string, i ...interface{}) {
	lg.ErrorField(fmt.Sprintf(strings.TrimRight(s, "\n"), i...), lg.Context(ctx))
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	cost := time.Since(begin)
	sql, rows := fc()
	fields := []lg.Field{
		lg.String("sql", sql),
		lg.Duration("cost", cost),
		lg.Int64("rows", rows),
		lg.Context(ctx),
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		lg.ErrorField("gorm", append(fields, lg.Err(err))...)
		return
	}
	if l.slowThreshold < cost && lg.Enable(lg.LvWarn) {
		lg.WarnField("gorm", fields...)
		return
	}
	if lg.Enable(lg.LvDebug) {
		lg.DebugField("gorm", fields...)
	}
	return
}

func (l *logger) LogMode(_ gormlog.LogLevel) gormlog.Interface {
	return l
}
