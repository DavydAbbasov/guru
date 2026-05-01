package metrics

import "go.uber.org/fx"

type Config struct {
	Namespace string // Prometheus metric namespace prefix
}

func NewFromConfig(cfg *Config) *Metrics {
	return New(cfg.Namespace)
}

var Module = fx.Module("metrics",
	fx.Provide(NewFromConfig),
)
