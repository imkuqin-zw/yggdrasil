package xgorm

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/imkuqin-zw/yggdrasil/contrib/gorm/driver"
	"github.com/imkuqin-zw/yggdrasil/contrib/gorm/plugin"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	lg "github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Open(conf *Config) *gorm.DB {
	conf.SetDefault()
	cfg := &gorm.Config{
		DryRun:                                   conf.DryRun,
		PrepareStmt:                              conf.PrepareStmt,
		DisableAutomaticPing:                     conf.DisableAutomaticPing,
		DisableForeignKeyConstraintWhenMigrating: conf.DisableForeignKeyConstraintWhenMigrating,
		IgnoreRelationshipsWhenMigrating:         conf.IgnoreRelationshipsWhenMigrating,
		DisableNestedTransaction:                 conf.DisableNestedTransaction,
		AllowGlobalUpdate:                        conf.AllowGlobalUpdate,
		QueryFields:                              conf.QueryFields,
		CreateBatchSize:                          conf.CreateBatchSize,
		TranslateError:                           conf.TranslateError,
		SkipDefaultTransaction:                   conf.SkipDefaultTransaction,
		FullSaveAssociations:                     conf.SkipDefaultTransaction,
		Logger: &logger{
			slowThreshold: conf.SlowThreshold,
		},
		NamingStrategy: schema.NamingStrategy{
			SingularTable: conf.NameStrategy.SingularTable,
			TablePrefix:   conf.NameStrategy.TablePrefix,
			NoLowerCase:   conf.NameStrategy.NoLowerCase,
		},
	}
	f := driver.GetFactory(conf.Driver)
	if f == nil {
		lg.FatalField("unknown gorm driver", lg.String("name", conf.Driver))
		return nil
	}
	dialector := f(conf.DSN)
	dsnCfg, err := dialector.ParseDSN(conf.DSN)
	if err != nil {
		lg.FatalField("fault to parse DSN", lg.Err(err))
		return nil
	}
	params := url.Values{}
	for k, v := range dsnCfg.Attrs {
		params.Add(k, v)
	}
	_ = config.SetMulti([]string{
		fmt.Sprintf("gorm.%s.addr", conf.instance),
		fmt.Sprintf("gorm.%s.user", conf.instance),
		fmt.Sprintf("gorm.%s.dbName", conf.instance),
		fmt.Sprintf("gorm.%s.params", conf.instance),
	}, []interface{}{
		dsnCfg.Addr,
		dsnCfg.User,
		dsnCfg.DBName,
		params.Encode(),
	})
	db, err := gorm.Open(dialector, cfg)
	if err != nil {
		lg.FatalField("fault to connect mysql", lg.Err(err))
		return nil
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil
	}
	sqlDb.SetMaxOpenConns(conf.MaxOpenConn)
	sqlDb.SetMaxIdleConns(conf.MaxIdleConn)
	sqlDb.SetConnMaxLifetime(conf.ConnMaxLifetime)
	sqlDb.SetConnMaxIdleTime(conf.ConnMaxIdleTime)

	for _, name := range conf.Plugins {
		if err = db.Use(plugin.GetPlugin(name, conf.instance)); err != nil {
			lg.FatalField("fault to use plugin", lg.Err(err))
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	if err := sqlDb.PingContext(ctx); err != nil {
		lg.FatalField("fault to ping mysql", lg.Err(err))
		return nil
	}
	return db
}

func NewDB(name string) *gorm.DB {
	c := new(Config)
	if err := config.Get("gorm." + name).Scan(c); err != nil {
		lg.FatalField("fault to load gorm config", lg.Err(err))
	}
	c.instance = name
	return Open(c)
}
