package metrics

import "go.uber.org/fx"

type Config struct {
	Namespace string
}

func NewFromConfig(cfg *Config) *Metrics {
	return New(cfg.Namespace)
}

var Module = fx.Module("metrics",
	fx.Provide(NewFromConfig),
)
