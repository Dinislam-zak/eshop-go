package repository

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const txKey contextKey = "tx"

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *transactionManager {
	return &transactionManager{db: db}
}

func (m *transactionManager) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}

func getTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return nil
}