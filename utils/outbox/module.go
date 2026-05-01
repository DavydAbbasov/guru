package outbox

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

var Module = fx.Module("outbox",
	fx.Provide(
		NewBuilder,
		NewKafkaPublisher,
		provideMetrics,
		NewDispatcher,
	),
	fx.Invoke(registerDispatcher),
)

func provideMetrics(reg *prometheus.Registry, cfg Config) *Metrics {
	return NewMetrics(reg, cfg.Namespace)
}

func registerDispatcher(lc fx.Lifecycle, d *Dispatcher) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go d.Run(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}
