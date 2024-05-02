package driver

import "gorm.io/gorm"

type DsnConfig struct {
	Addr   string
	User   string
	DBName string
	Attrs  map[string]string
}

type Dialector interface {
	gorm.Dialector
	ParseDSN(string) (*DsnConfig, error)
}

var driverFactory = make(map[string]func(string) Dialector)

func RegisterFactory(name string, f func(string) Dialector) {
	driverFactory[name] = f
}

func GetFactory(name string) func(string) Dialector {
	f, _ := driverFactory[name]
	return f
}
