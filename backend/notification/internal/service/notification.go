package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"

	"guru/backend/notification/internal/entities"
	notificationmetrics "guru/backend/notification/internal/metrics"
	"guru/backend/notification/internal/notifier"
	"guru/utils/logger"
	"guru/utils/tracing"
)

var ErrInvalidEvent = errors.New("invalid event")

type NotificationService struct {
	notifier notifier.Notifier
	metrics  *notificationmetrics.Metrics
	tracer   *tracing.Tracer
	log      logger.Logger
}

func NewNotificationService(
	n notifier.Notifier,
	m *notificationmetrics.Metrics,
	tracer *tracing.Tracer,
	log logger.Logger,
) *NotificationService {
	return &NotificationService{
		notifier: n,
		metrics:  m,
		tracer:   tracer,
		log:      log,
	}
}

func (s *NotificationService) Process(ctx context.Context, event *entities.ProductEvent) error {
	ctx, span := s.tracer.Start(ctx, "NotificationService.Process")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	eventType := event.Type
	if eventType == "" {
		eventType = "unknown"
	}

	s.metrics.EventsConsumed.WithLabelValues(eventType).Inc()
	start := time.Now()
	defer func() {
		s.metrics.ProcessingDuration.WithLabelValues(eventType).Observe(time.Since(start).Seconds())
	}()

	if err := validate(event); err != nil {
		s.metrics.EventsFailed.WithLabelValues(eventType, "validation").Inc()
		log.Warn("invalid event; skipping",
			zap.String("event_id", event.ID),
			zap.Error(err))
		return err
	}

	if err := s.notifier.Notify(ctx, event); err != nil {
		s.metrics.EventsFailed.WithLabelValues(eventType, "notify").Inc()
		log.Error("notify failed",
			zap.String("event_id", event.ID),
			zap.Error(err))
		return err
	}

	s.metrics.EventsProcessed.WithLabelValues(eventType).Inc()
	return nil
}

func validate(e *entities.ProductEvent) error {
	if e == nil || strings.TrimSpace(e.ID) == "" {
		return ErrInvalidEvent
	}
	if strings.TrimSpace(e.Type) == "" {
		return ErrInvalidEvent
	}
	return nil
}
