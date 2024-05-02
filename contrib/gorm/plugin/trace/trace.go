package trace

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"

	"github.com/imkuqin-zw/yggdrasil/contrib/gorm/plugin"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.1"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

func init() {
	plugin.RegisterPluginFactory("trace", newPlugin)
}

var dbRowsAffected = attribute.Key("db.rows_affected")

type tracePlugin struct {
	tracer           trace.Tracer
	attrs            []attribute.KeyValue
	excludeQueryVars bool
	queryFormatter   func(query string) string
}

func (p *tracePlugin) Name() string {
	return "trace"
}

func newPlugin(instance string) gorm.Plugin {
	cfg := &Config{}
	if err := config.Get("gorm." + instance + ".trace").Scan(cfg); err != nil {
		logger.ErrorField("fault to load gorm trace config", logger.Err(err), logger.String("name", instance))
		return nil
	}
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
	p := &tracePlugin{
		tracer: otel.Tracer("github.com/imkuqin-zw/yggdrasil/contrib/gorm/plugin/trace",
			trace.WithInstrumentationVersion("semver:1.25.7")),
		excludeQueryVars: cfg.ExcludeQueryVars,
		queryFormatter:   nil,
		attrs: []attribute.KeyValue{
			semconv.DBNameKey.String(instanceInfo.DBName),
			semconv.DBUser(instanceInfo.User),
			semconv.DBConnectionString(connStr),
		},
	}
	return p
}

type gormHookFunc func(tx *gorm.DB)

type gormRegister interface {
	Register(name string, fn func(*gorm.DB)) error
}

func (p *tracePlugin) Initialize(db *gorm.DB) (err error) {
	cb := db.Callback()
	hooks := []struct {
		callback gormRegister
		hook     gormHookFunc
		name     string
	}{
		{cb.Create().Before("gorm:create"), p.before("gorm.Create"), "before:create"},
		{cb.Create().After("gorm:create"), p.after(), "after:create"},

		{cb.Query().Before("gorm:query"), p.before("gorm.Query"), "before:select"},
		{cb.Query().After("gorm:query"), p.after(), "after:select"},

		{cb.Delete().Before("gorm:delete"), p.before("gorm.Delete"), "before:delete"},
		{cb.Delete().After("gorm:delete"), p.after(), "after:delete"},

		{cb.Update().Before("gorm:update"), p.before("gorm.Update"), "before:update"},
		{cb.Update().After("gorm:update"), p.after(), "after:update"},

		{cb.Row().Before("gorm:row"), p.before("gorm.Row"), "before:row"},
		{cb.Row().After("gorm:row"), p.after(), "after:row"},

		{cb.Raw().Before("gorm:raw"), p.before("gorm.Raw"), "before:raw"},
		{cb.Raw().After("gorm:raw"), p.after(), "after:raw"},
	}

	var firstErr error

	for _, h := range hooks {
		if err = h.callback.Register("trace:"+h.name, h.hook); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("callback register %s failed: %w", h.name, err)
		}
	}

	return firstErr
}

func (p *tracePlugin) before(spanName string) gormHookFunc {
	return func(tx *gorm.DB) {
		tx.Statement.Context, _ = p.tracer.Start(tx.Statement.Context, spanName, trace.WithSpanKind(trace.SpanKindClient))
	}
}

func (p *tracePlugin) after() gormHookFunc {
	return func(tx *gorm.DB) {
		span := trace.SpanFromContext(tx.Statement.Context)
		if !span.IsRecording() {
			return
		}
		defer span.End()

		attrs := make([]attribute.KeyValue, 0, len(p.attrs)+4)
		attrs = append(attrs, p.attrs...)

		if sys := dbSystem(tx); sys.Valid() {
			attrs = append(attrs, sys)
		}

		vars := tx.Statement.Vars

		var query string
		if p.excludeQueryVars {
			query = tx.Statement.SQL.String()
		} else {
			query = tx.Dialector.Explain(tx.Statement.SQL.String(), vars...)
		}
		attrs = append(attrs, semconv.DBStatementKey.String(p.formatQuery(query)))
		if tx.Statement.Table != "" {
			attrs = append(attrs, semconv.DBSQLTableKey.String(tx.Statement.Table))
		}
		if tx.Statement.RowsAffected != -1 {
			attrs = append(attrs, dbRowsAffected.Int64(tx.Statement.RowsAffected))
		}

		span.SetAttributes(attrs...)
		switch tx.Error {
		case nil,
			gorm.ErrRecordNotFound,
			driver.ErrSkip,
			io.EOF, // end of rows iterator
			sql.ErrNoRows:
			// ignore
		default:
			span.RecordError(tx.Error)
			span.SetStatus(codes.Error, tx.Error.Error())
		}
	}
}

func (p *tracePlugin) formatQuery(query string) string {
	if p.queryFormatter != nil {
		return p.queryFormatter(query)
	}
	return query
}

func dbSystem(tx *gorm.DB) attribute.KeyValue {
	switch tx.Dialector.Name() {
	case "mysql":
		return semconv.DBSystemMySQL
	case "mssql":
		return semconv.DBSystemMSSQL
	case "postgres", "postgresql":
		return semconv.DBSystemPostgreSQL
	case "sqlite":
		return semconv.DBSystemSqlite
	case "sqlserver":
		return semconv.DBSystemKey.String("sqlserver")
	case "clickhouse":
		return semconv.DBSystemKey.String("clickhouse")
	default:
		return attribute.KeyValue{}
	}
}
