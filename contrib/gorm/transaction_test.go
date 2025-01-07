package xgorm

import (
	"context"
	"testing"

	_ "github.com/imkuqin-zw/yggdrasil/contrib/gorm/driver/sqlite"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_transaction(t *testing.T) {
	_ = config.Set("gorm.test.dsn", ":memory:")
	_ = config.Set("gorm.test.driver", "sqlite")
	config.Set("gorm.test.nameStrategy.singularTable", true)
	db := NewDB("test")
	require.NoError(t, db.Exec("create table test(id int primary key,name varchar(255))").Error)

	trans := NewTransaction(db)
	ctx := trans.Begin(context.TODO())
	require.Equal(t, ctx, trans.Begin(ctx))
	db.WithContext(ctx)
	if !assert.NoError(t, trans.GetTx(ctx).Exec("insert into test(id,name) values(1,'test')").Error) {
		_ = trans.Rollback(ctx)
		return
	}
	if !assert.NoError(t, trans.GetTx(ctx).Exec("insert into test(id,name) values(2,'test2')").Error) {
		_ = trans.Rollback(ctx)
		return
	}
	tx1 := trans.GetTx(ctx)
	tx2 := trans.GetTx(ctx)
	require.Equal(t, tx1, tx2)
	require.NoError(t, trans.Commit(ctx))
	require.NoError(t, trans.Rollback(ctx))
	tx3 := trans.GetTx(ctx)
	require.NotEqual(t, tx1, tx3)

	var user = struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}{}
	err := trans.GetTx(ctx).Table("test").Where("id=?", 1).First(&user).Error
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, user.Id)

	require.NotEqual(t, ctx, trans.Begin(ctx))
}
