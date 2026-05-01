package metrics

import (
	"guru/backend/notification/internal/config"
	utilsMetrics "guru/utils/metrics"

	"go.uber.org/fx"
)

var Module = fx.Module("notification-metrics",
	fx.Provide(func(
		parent *utilsMetrics.Metrics,
		cfg *config.MetricsConfig,
	) *Metrics {
		return New(parent.Registry, cfg.Namespace)
	}),
)
