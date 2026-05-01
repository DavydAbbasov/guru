package config

import "go.uber.org/fx"

var Module = fx.Module("config",
	fx.Provide(
		New,
		func(cfg *Config) *ServerConfig  { return cfg.Server },
		func(cfg *Config) *KafkaConfig   { return cfg.Kafka },
		func(cfg *Config) *LoggerConfig  { return cfg.Logger },
		func(cfg *Config) *TracerConfig  { return cfg.Tracer },
		func(cfg *Config) *MetricsConfig { return cfg.Metrics },
	),
)
