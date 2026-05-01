package metrics

import (
	"guru/backend/products/internal/config"
	utilsMetrics "guru/utils/metrics"

	"go.uber.org/fx"
)

var Module = fx.Module("products-metrics",
	fx.Provide(func(
		parent *utilsMetrics.Metrics,
		cfg *config.MetricsConfig,
	) *Metrics {
		return New(parent.Registry, cfg.Namespace)
	}),
)
