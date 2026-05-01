package container

import (
	"guru/backend/notification/internal/config"
	notificationmetrics "guru/backend/notification/internal/metrics"
	"guru/backend/notification/internal/notifier"
	"guru/backend/notification/internal/service"
	httptransport "guru/backend/notification/internal/transport/http"
	kafkatransport "guru/backend/notification/internal/transport/kafka"
	notificationpgsql "guru/backend/notification/pkg/pgsql"
	"guru/utils/idempotency"
	"guru/utils/logger"
	"guru/utils/metrics"
	"guru/utils/pgsql"
	"guru/utils/profiling"
	"guru/utils/tracing"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

func Build() *fx.App {
	return fx.New(
		config.Module,
		fx.Provide(
			adaptLoggerConfig,
			adaptTracerConfig,
			adaptMetricsConfig,
			adaptProfilingConfig,
			adaptDBConfig,
			adaptIdempotencyConfig,
			adaptPrometheusRegistry,
		),
		logger.Module,
		tracing.Module,
		metrics.Module,
		notificationpgsql.Module,

		notificationmetrics.Module,
		notifier.Module,
		idempotency.Module,
		service.Module,

		kafkatransport.Module,
		httptransport.Module,
		profiling.Module,

		fx.Invoke(pgsql.RegisterMetrics),
	)
}

func adaptLoggerConfig(cfg *config.LoggerConfig, tcfg *config.TracerConfig) *logger.Config {
	return &logger.Config{Level: cfg.Level, Service: tcfg.ServiceName}
}

func adaptTracerConfig(cfg *config.TracerConfig) *tracing.Config {
	return &tracing.Config{
		Disabled:    cfg.Disabled,
		Endpoint:    cfg.Endpoint,
		ServiceName: cfg.ServiceName,
		Insecure:    true,
	}
}

func adaptMetricsConfig(cfg *config.MetricsConfig) *metrics.Config {
	return &metrics.Config{Namespace: cfg.Namespace}
}

func adaptDBConfig(cfg *config.DatabaseConfig) *pgsql.Config {
	return &pgsql.Config{
		Host: cfg.Host,
		Port: cfg.Port,
		User: cfg.User,
		Pass: cfg.Pass,
		Name: cfg.Name,
	}
}

func adaptIdempotencyConfig(mcfg *config.MetricsConfig) idempotency.Config {
	return idempotency.Config{Namespace: mcfg.Namespace}
}

func adaptPrometheusRegistry(m *metrics.Metrics) *prometheus.Registry {
	return m.Registry
}

func adaptProfilingConfig(cfg *config.ProfilingConfig, tcfg *config.TracerConfig) *profiling.Config {
	path := cfg.Path
	if path == "" {
		path = "/var/log/goprofile"
	}
	return &profiling.Config{
		CPU:         cfg.CPU,
		Memory:      cfg.Memory,
		Path:        path,
		ServiceName: tcfg.ServiceName,
	}
}
