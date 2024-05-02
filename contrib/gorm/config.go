package xgorm

import (
	"fmt"
	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"time"
)

const (
	defaultMaxIdleConn     = 10
	defaultMaxOpenConn     = 100
	defaultConnMaxLifetime = time.Second * 300
	defaultSlowThreshold   = time.Millisecond * 500
)

type Config struct {
	instance string
	// DSN: user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	DSN string
	// Driver  the name of driver
	Driver string
	// DryRun generate sql without execute
	DryRun bool
	// PrepareStmt executes the given query in cached statement
	PrepareStmt bool
	// DisableAutomaticPing
	DisableAutomaticPing bool
	// DisableForeignKeyConstraintWhenMigrating
	DisableForeignKeyConstraintWhenMigrating bool
	// IgnoreRelationshipsWhenMigrating
	IgnoreRelationshipsWhenMigrating bool
	// DisableNestedTransaction disable nested transaction
	DisableNestedTransaction bool
	// AllowGlobalUpdate allow global update
	AllowGlobalUpdate bool
	// QueryFields executes the SQL query with all fields of the table
	QueryFields bool
	// CreateBatchSize default create batch size
	CreateBatchSize int
	// TranslateError enabling error translation
	TranslateError bool
	// GORM perform single create, update, delete operations in transactions by default to ensure database data integrity
	// You can disable it by setting `SkipDefaultTransaction` to true
	SkipDefaultTransaction bool
	// FullSaveAssociations full save associations
	FullSaveAssociations bool
	// MaxIdleConn the maximum number of connections in the idle
	// connection pool.
	MaxIdleConn int
	MaxOpenConn int
	// ConnMaxLifetime the maximum amount of time a connection may be reused.
	ConnMaxLifetime time.Duration
	//  ConnMaxIdleTime the maximum amount of time a connection may be idle.
	ConnMaxIdleTime time.Duration
	NameStrategy    struct {
		TablePrefix   string
		SingularTable bool
		NoLowerCase   bool
	}
	SlowThreshold time.Duration
	Plugins       []string
}

// Check
func (config *Config) SetDefault() {
	if config.Driver == "" {
		config.Driver = "mysql"
		_ = config2.Set(fmt.Sprintf("gorm.%s.driver", config.instance), config.Driver)
	}
	if config.MaxIdleConn == 0 {
		config.MaxIdleConn = defaultMaxIdleConn
	}
	if config.MaxOpenConn == 0 {
		config.MaxOpenConn = defaultMaxOpenConn
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = defaultConnMaxLifetime
	}
	if config.SlowThreshold < time.Millisecond {
		config.SlowThreshold = defaultSlowThreshold
	}
}
