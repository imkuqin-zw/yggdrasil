package mysql

import (
	sqlMysql "github.com/go-sql-driver/mysql"
	"github.com/imkuqin-zw/yggdrasil/contrib/gorm/driver"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	driver.RegisterFactory("mysql", NewDialector)
}

func NewDialector(dsn string) driver.Dialector {
	return &Dialector{
		Dialector: mysql.Open(dsn),
	}
}

type Dialector struct {
	gorm.Dialector
}

func (d *Dialector) ParseDSN(dsn string) (*driver.DsnConfig, error) {
	cfg, err := sqlMysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return &driver.DsnConfig{
		Addr:   cfg.Addr,
		User:   cfg.User,
		DBName: cfg.DBName,
		Attrs:  cfg.Params,
	}, nil
}
