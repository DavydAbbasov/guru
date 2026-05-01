package pgsql

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Module("pgsql",
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
			}
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("failed to ping database: %w", err)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
			}
			return sqlDB.Close()
		},
	})
}
