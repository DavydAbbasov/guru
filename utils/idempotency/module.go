package idempotency

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

var Module = fx.Module("idempotency",
	fx.Provide(
		NewRepository,
		provideMetrics,
		NewCleanup,
	),
	fx.Invoke(registerCleanup),
)

func provideMetrics(reg *prometheus.Registry, cfg Config) *Metrics {
	return NewMetrics(reg, cfg.Namespace)
}

func registerCleanup(lc fx.Lifecycle, c *Cleanup) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go c.Run(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}
