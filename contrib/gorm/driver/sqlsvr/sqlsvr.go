package sqlsvr

import (
	"fmt"
	"github.com/imkuqin-zw/yggdrasil/contrib/gorm/driver"
	"github.com/microsoft/go-mssqldb/msdsn"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func init() {
	driver.RegisterFactory("sqlserver", NewDialector)
}

func NewDialector(dsn string) driver.Dialector {
	return &Dialector{
		Dialector: sqlserver.Open(dsn),
	}
}

type Dialector struct {
	gorm.Dialector
}

func (d *Dialector) ParseDSN(dsn string) (*driver.DsnConfig, error) {
	cfg, err := msdsn.Parse(dsn)
	if err != nil {
		return nil, err
	}
	return &driver.DsnConfig{
		Addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:   cfg.User,
		DBName: cfg.Database,
		Attrs:  cfg.Parameters,
	}, nil
}
