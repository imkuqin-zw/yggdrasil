package metrics

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/contrib/gorm/plugin"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.1"
	"gorm.io/gorm"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func init() {
	plugin.RegisterPluginFactory("metrics", newPlugin)
}

type metricPlugin struct {
	meter metric.Meter
	attrs []attribute.KeyValue
}

func (p *metricPlugin) Name() string {
	return "metrics"
}

func newPlugin(instance string) gorm.Plugin {
	var instanceInfo = struct {
		Addr   string `json:"addr"`
		User   string `json:"user"`
		DBName string `json:"dbName"`
		Driver string `json:"driver"`
		Params string `json:"params"`
	}{}
	if err := config.Get("gorm." + instance).Scan(&instanceInfo); err != nil {
		logger.FatalField("fault to get gorm instance info", logger.Err(err), logger.String("name", instance))
	}
	connStr := fmt.Sprintf("%s://%s", instanceInfo.Driver, instanceInfo.Addr)
	if len(instanceInfo.Params) > 0 {
		connStr = fmt.Sprintf("%s?%s", connStr, instanceInfo.Params)
	}
	p := &metricPlugin{
		meter: otel.Meter("github.com/imkuqin-zw/yggdrasil/contrib/gorm/plugin/metrics",
			metric.WithInstrumentationVersion("semver:1.25.7")),
		attrs: []attribute.KeyValue{
			semconv.DBNameKey.String(instanceInfo.DBName),
			semconv.DBUser(instanceInfo.User),
			semconv.DBConnectionString(connStr),
			semconv.DBSystemKey.String(instanceInfo.Driver),
		},
	}
	return p
}

func (p *metricPlugin) Initialize(db *gorm.DB) (err error) {
	sqldb, ok := db.ConnPool.(*sql.DB)
	if !ok {
		logger.Warn("the connection type does not support metrics monitoring")
		return nil
	}
	maxOpenConns, _ := p.meter.Int64ObservableGauge(
		"sql.connections_max_open",
		metric.WithDescription("Maximum number of open connections to the database"),
	)
	openConns, _ := p.meter.Int64ObservableGauge(
		"sql.connections_open",
		metric.WithDescription("The number of established connections both in use and idle"),
	)
	connsWaitCount, _ := p.meter.Int64ObservableCounter(
		"sql.connections_wait_count",
		metric.WithDescription("The total number of connections waited for"),
	)
	connsWaitDuration, _ := p.meter.Int64ObservableCounter(
		"sql.connections_wait_duration",
		metric.WithDescription("The total time blocked waiting for a new connection"),
		metric.WithUnit("nanoseconds"),
	)
	connsClosedMaxIdle, _ := p.meter.Int64ObservableCounter(
		"sql.connections_closed_max_idle",
		metric.WithDescription("The total number of connections closed due to SetMaxIdleConns"),
	)
	connsClosedMaxIdleTime, _ := p.meter.Int64ObservableCounter(
		"sql.connections_closed_max_idle_time",
		metric.WithDescription("The total number of connections closed due to SetConnMaxIdleTime"),
	)
	connsClosedMaxLifetime, _ := p.meter.Int64ObservableCounter(
		"sql.connections_closed_max_lifetime",
		metric.WithDescription("The total number of connections closed due to SetConnMaxLifetime"),
	)
	idleAttrs := append(p.attrs, attribute.String("state", "idle"))
	usedAttrs := append(p.attrs, attribute.String("state", "used"))

	_, err = p.meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			stats := sqldb.Stats()
			o.ObserveInt64(maxOpenConns, int64(stats.MaxOpenConnections), metric.WithAttributes(p.attrs...))
			o.ObserveInt64(openConns, int64(stats.InUse), metric.WithAttributes(usedAttrs...))
			o.ObserveInt64(openConns, int64(stats.Idle), metric.WithAttributes(idleAttrs...))
			o.ObserveInt64(connsWaitCount, stats.WaitCount, metric.WithAttributes(p.attrs...))
			o.ObserveInt64(connsWaitDuration, int64(stats.WaitDuration), metric.WithAttributes(p.attrs...))
			o.ObserveInt64(connsClosedMaxIdle, stats.MaxIdleClosed, metric.WithAttributes(p.attrs...))
			o.ObserveInt64(connsClosedMaxIdleTime, stats.MaxIdleTimeClosed, metric.WithAttributes(p.attrs...))
			o.ObserveInt64(connsClosedMaxLifetime, stats.MaxLifetimeClosed, metric.WithAttributes(p.attrs...))
			return nil
		},
		maxOpenConns,
		openConns,
		connsWaitCount,
		connsWaitDuration,
		connsClosedMaxIdle,
		connsClosedMaxIdleTime,
		connsClosedMaxLifetime,
	)
	if err != nil {
		return err
	}
	return nil
}
