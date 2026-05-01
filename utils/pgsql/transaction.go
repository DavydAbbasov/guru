package pgsql

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TransactionManager struct {
	DB            *gorm.DB
	commitTimeout time.Duration
}

func NewTransactionManager(db *gorm.DB, commitTimeout time.Duration) *TransactionManager {
	if commitTimeout <= 0 {
		commitTimeout = 10 * time.Second
	}
	return &TransactionManager{DB: db, commitTimeout: commitTimeout}
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := tm.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("transaction error: %w (rollback: %v)", err, rbErr)
		}
		return err
	}

	// commit must finish even if the caller's ctx was cancelled mid-tx; otherwise the open tx leaks.
	commitCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), tm.commitTimeout)
	defer cancel()

	if err := tx.WithContext(commitCtx).Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (tm *TransactionManager) WithoutTransaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return fn(tm.DB.WithContext(ctx))
}
