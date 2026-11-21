package database

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager interface {
    Transaction(
        ctx context.Context,
        fn func(tx *gorm.DB) error,
    ) error
}

type postgresTransactionManager struct {
    db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
    return &postgresTransactionManager{
        db: db,
    }
}

func (tm *postgresTransactionManager) Transaction(
    ctx context.Context,
    fn func(tx *gorm.DB) error,
) error {
    return tm.db.WithContext(ctx).Transaction(fn)
}