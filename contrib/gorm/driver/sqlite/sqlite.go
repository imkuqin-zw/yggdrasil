package sqlite

import (
	"net/url"
	"strings"

	"github.com/imkuqin-zw/yggdrasil/contrib/gorm/driver"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	driver.RegisterFactory("sqlite", NewDialector)
}

func NewDialector(dsn string) driver.Dialector {
	return &Dialector{
		Dialector: sqlite.Open(dsn),
	}
}

type Dialector struct {
	gorm.Dialector
}

func (d *Dialector) ParseDSN(dsn string) (*driver.DsnConfig, error) {
	dsnParts := strings.SplitN(dsn, "?", 2)
	cfg := &driver.DsnConfig{
		Addr:   dsnParts[0],
		DBName: dsnParts[0],
	}
	if len(dsnParts) > 1 {
		params, _ := url.ParseQuery(dsnParts[1])
		cfg.Attrs = make(map[string]string, len(params))
		for key, values := range params {
			cfg.Attrs[key] = values[0]
		}
	} else {
		cfg.Attrs = make(map[string]string)
	}
	return cfg, nil
}
