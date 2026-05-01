package container

import (
	"guru/backend/products/internal/config"
	productsmetrics "guru/backend/products/internal/metrics"
	pgsqlrepo "guru/backend/products/internal/repository/pgsql"
	"guru/backend/products/internal/service"
	httptransport "guru/backend/products/internal/transport/http"
	productspgsql "guru/backend/products/pkg/pgsql"
	kafkatool "guru/utils/kafka-tool"
	"guru/utils/logger"
	"guru/utils/metrics"
	"guru/utils/outbox"
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
			adaptDBConfig,
			adaptKafkaConfig,
			adaptProfilingConfig,
			adaptOutboxConfig,
			adaptPrometheusRegistry,
		),

		logger.Module,
		tracing.Module,
		metrics.Module,
		productspgsql.Module,
		kafkatool.ProducerModule,

		productsmetrics.Module,
		pgsqlrepo.Module,
		outbox.Module,
		service.Module,

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

func adaptKafkaConfig(cfg *config.KafkaConfig) *kafkatool.Config {
	return &kafkatool.Config{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		ClientID: "products-service",
	}
}

func adaptOutboxConfig(mcfg *config.MetricsConfig) outbox.Config {
	return outbox.Config{Namespace: mcfg.Namespace}
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
