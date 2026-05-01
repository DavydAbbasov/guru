package config

import "go.uber.org/fx"

var Module = fx.Module("config",
	fx.Provide(
		New,
		func(cfg *Config) *ServerConfig   { return cfg.Server },
		func(cfg *Config) *DatabaseConfig { return cfg.Database },
		func(cfg *Config) *KafkaConfig    { return cfg.Kafka },
		func(cfg *Config) *LoggerConfig   { return cfg.Logger },
		func(cfg *Config) *TracerConfig   { return cfg.Tracer },
		func(cfg *Config) *MetricsConfig  { return cfg.Metrics },
		func(cfg *Config) *ProfilingConfig {
			if cfg.Profiling == nil {
				return &ProfilingConfig{}
			}
			return cfg.Profiling
		},
		func(cfg *Config) *OutboxConfig {
			if cfg.Outbox == nil {
				return &OutboxConfig{}
			}
			return cfg.Outbox
		},
	),
)
