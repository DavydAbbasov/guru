package logger

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Config struct {
	Level   string
	Service string
}

func NewLoggerFromConfig(cfg *Config) (Logger, error) {
	lvl, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	log, err := NewZapLogger(lvl.Level())
	if err != nil {
		return nil, err
	}
	if cfg.Service != "" {
		log = log.With(zap.String("service", cfg.Service))
	}
	return log, nil
}

var Module = fx.Module("logger",
	fx.Provide(NewLoggerFromConfig),
	fx.Invoke(registerLoggerLifecycle),
)

func registerLoggerLifecycle(lc fx.Lifecycle, l Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = l.Sync() // sync errors on stderr/stdout are expected
			return nil
		},
	})
}
