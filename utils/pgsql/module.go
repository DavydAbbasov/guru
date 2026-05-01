package pgsql

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = ModuleWithOptions()

func ModuleWithOptions(optFuncs ...OptionFunc) fx.Option {
	opts := defaultOptions()
	for _, fn := range optFuncs {
		fn(opts)
	}

	return fx.Module("pgsql",
		fx.Provide(
			func(cfg *Config) (*gorm.DB, error) { return New(cfg, optFuncs...) },
			func(db *gorm.DB) *TransactionManager { return NewTransactionManager(db, opts.CommitTimeout) },
		),
		fx.Invoke(registerLifecycle),
	)
}

func registerLifecycle(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("get underlying *sql.DB: %w", err)
			}
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("ping database: %w", err)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("get underlying *sql.DB: %w", err)
			}
			return sqlDB.Close()
		},
	})
}
