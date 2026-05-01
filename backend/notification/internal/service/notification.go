package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"guru/backend/notification/internal/entities"
	notificationmetrics "guru/backend/notification/internal/metrics"
	"guru/backend/notification/internal/notifier"
	"guru/utils/logger"
	"guru/utils/tracing"
)

var ErrInvalidEvent = errors.New("invalid event")

type IdempotencyRepository interface {
	MarkProcessed(ctx context.Context, id uuid.UUID, eventType string) (firstTime bool, err error)
}

type NotificationService struct {
	notifier notifier.Notifier
	idem     IdempotencyRepository
	metrics  *notificationmetrics.Metrics
	tracer   *tracing.Tracer
	log      logger.Logger
}

func NewNotificationService(
	n notifier.Notifier,
	idem IdempotencyRepository,
	m *notificationmetrics.Metrics,
	tracer *tracing.Tracer,
	log logger.Logger,
) *NotificationService {
	return &NotificationService{
		notifier: n,
		idem:     idem,
		metrics:  m,
		tracer:   tracer,
		log:      log,
	}
}

func (s *NotificationService) Process(ctx context.Context, event *entities.ProductEvent, eventID string) error {
	ctx, span := s.tracer.Start(ctx, "NotificationService.Process")
	defer span.End()

	log := logger.FromContext(ctx, s.log)
	eventType := "unknown"
	if event != nil && event.Type != "" {
		eventType = event.Type
	}

	s.metrics.EventsConsumed.WithLabelValues(eventType).Inc()
	start := time.Now()
	defer func() {
		s.metrics.ProcessingDuration.WithLabelValues(eventType).Observe(time.Since(start).Seconds())
	}()

	if err := validate(event); err != nil {
		s.metrics.EventsFailed.WithLabelValues(eventType, "validation").Inc()
		log.Warn("invalid event; skipping",
			zap.String("event_id", eventIDStr(event)),
			zap.Error(err))
		return err
	}

	if eventID != "" {
		id, err := uuid.Parse(eventID)
		if err != nil {
			s.metrics.EventsFailed.WithLabelValues(eventType, "bad_event_id").Inc()
			log.Warn("invalid outbox-event-id header; skipping dedup",
				zap.String("outbox_event_id", eventID),
				zap.Error(err))
		} else {
			first, err := s.idem.MarkProcessed(ctx, id, eventType)
			if err != nil {
				s.metrics.EventsFailed.WithLabelValues(eventType, "dedup").Inc()
				log.Error("dedup check failed",
					zap.String("outbox_event_id", eventID),
					zap.Error(err))
				return err
			}
			if !first {
				log.Info("duplicate event skipped",
					zap.String("outbox_event_id", eventID),
					zap.String("event_id", event.ID))
				return nil
			}
		}
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

func eventIDStr(e *entities.ProductEvent) string {
	if e == nil {
		return ""
	}
	return e.ID
}
