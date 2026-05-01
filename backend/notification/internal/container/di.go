package container

import (
	"guru/backend/notification/internal/config"
	notificationmetrics "guru/backend/notification/internal/metrics"
	"guru/backend/notification/internal/notifier"
	"guru/backend/notification/internal/service"
	httptransport "guru/backend/notification/internal/transport/http"
	kafkatransport "guru/backend/notification/internal/transport/kafka"
	"guru/utils/logger"
	"guru/utils/metrics"
	"guru/utils/tracing"

	"go.uber.org/fx"
)

func Build() *fx.App {
	return fx.New(
		config.Module,
		fx.Provide(
			adaptLoggerConfig,
			adaptTracerConfig,
			adaptMetricsConfig,
		),
		logger.Module,
		tracing.Module,
		metrics.Module,

		notificationmetrics.Module,
		notifier.Module,
		service.Module,

		kafkatransport.Module,
		httptransport.Module,
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
