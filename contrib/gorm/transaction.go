package xgorm

import (
	"context"

	"gorm.io/gorm"
)

type Transaction interface {
	Begin(ctx context.Context) context.Context
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	GetTx(ctx context.Context) *gorm.DB
}

type transaction struct {
	db *gorm.DB
}

func NewTransaction(db *gorm.DB) Transaction {
	trans := &transaction{
		db: db,
	}
	return trans
}

func (t *transaction) Begin(ctx context.Context) context.Context {
	if tx, ok := ctx.Value("tx").(**gorm.DB); ok && *tx != nil {
		return ctx
	}
	tx := t.db.WithContext(ctx).Begin()
	return context.WithValue(ctx, "tx", &tx)
}

func (t *transaction) Commit(ctx context.Context) error {
	if tx, ok := ctx.Value("tx").(**gorm.DB); ok && *tx != nil {
		err := (*tx).WithContext(ctx).Commit().Error
		*tx = nil
		return err
	}
	return nil
}

func (t *transaction) Rollback(ctx context.Context) error {
	if tx, ok := ctx.Value("tx").(**gorm.DB); ok && *tx != nil {
		err := (*tx).WithContext(ctx).Rollback().Error
		*tx = nil
		return err
	}
	return nil
}

func (t *transaction) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(**gorm.DB); ok && *tx != nil {
		return *tx
	}
	return t.db
}
