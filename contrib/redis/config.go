package xredis

import "github.com/redis/go-redis/v9"

type Config struct {
	Universal     redis.UniversalOptions
	Cluster       bool
	MetricsEnable bool
	TraceEnable   bool
}
