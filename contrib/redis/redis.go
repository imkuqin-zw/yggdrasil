package xredis

import (
	"context"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type Redis interface {
	redis.UniversalClient
}

func NewRedis(name string) Redis {
	cfg := new(Config)
	if err := config.Get("redis." + name).Scan(cfg); err != nil {
		logger.FatalField("fault to load redis config", logger.Err(err))
	}
	redis.SetLogger(&logging{})
	cli := newUniversalClient(cfg)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		logger.FatalField("fault to connect redis", logger.Err(err))
	}
	return cli
}

func newUniversalClient(cfg *Config) redis.UniversalClient {
	if cfg.Universal.MasterName != "" {
		return redis.NewFailoverClient(cfg.Universal.Failover())
	}
	if cfg.Cluster || len(cfg.Universal.Addrs) > 1 {
		return redis.NewClusterClient(cfg.Universal.Cluster())
	}
	cli := redis.NewClient(cfg.Universal.Simple())
	InstrumentMetrics(cfg, cli)
	InstrumentTracing(cfg, cli)
	return cli
}

func InstrumentMetrics(cfg *Config, cli redis.UniversalClient) {
	if !cfg.MetricsEnable {
		return
	}

	if err := redisotel.InstrumentMetrics(cli); err != nil {
		logger.FatalField("fault to init  redis metrics", logger.Err(err))
	}
}

func InstrumentTracing(cfg *Config, cli redis.UniversalClient) {
	if !cfg.TraceEnable {
		return
	}
	if err := redisotel.InstrumentTracing(cli); err != nil {
		logger.FatalField("fault to init  redis trace", logger.Err(err))
	}
}
