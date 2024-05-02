package xgorm

import (
	"testing"

	_ "github.com/imkuqin-zw/yggdrasil/contrib/gorm/driver/mysql"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/stretchr/testify/assert"
)

type Client struct {
	ID           uint32
	ClientID     string
	ClientSecret string
}

func TestNewDB(t *testing.T) {
	config.Set("gorm.test.driver", "mysql")
	config.Set("gorm.test.nameStrategy.singularTable", true)
	config.Set("gorm.test.dsn", "root:12345678@tcp(127.0.0.1)/uuid?charset=utf8&parseTime=True&loc=Local")
	db := NewDB("test")
	client := &Client{}
	b, err := FindOne(db, client)
	assert.True(t, b)
	assert.Nil(t, err)
}
