package kafka

import (
	"context"

	eventsv1 "guru/apis/products/v1/events"
	"guru/backend/notification/internal/entities"
	notificationmetrics "guru/backend/notification/internal/metrics"
	"guru/backend/notification/internal/service"
	"guru/utils/logger"
	"guru/utils/tracing"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ConsumerHandler struct {
	service *service.NotificationService
	metrics *notificationmetrics.Metrics
	log     logger.Logger
	tracer  *tracing.Tracer
}

func NewConsumerHandler(
	svc *service.NotificationService,
	m *notificationmetrics.Metrics,
	log logger.Logger,
	tracer *tracing.Tracer,
) *ConsumerHandler {
	return &ConsumerHandler{
		service: svc,
		metrics: m,
		log:     log.With(zap.String("component", "notification.kafka.handler")),
		tracer:  tracer,
	}
}

func (h *ConsumerHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage) error {
	ctx, span := h.tracer.Start(ctx, "ConsumerHandler.Handle",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.source", msg.Topic),
			attribute.String("messaging.operation", "process"),
			attribute.Int64("messaging.kafka.partition", int64(msg.Partition)),
			attribute.Int64("messaging.kafka.offset", msg.Offset),
			attribute.Int("messaging.message_payload_size_bytes", len(msg.Value)),
		))
	defer span.End()

	log := logger.FromContext(ctx, h.log).With(
		zap.String("topic", msg.Topic),
		zap.Int32("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
	)

	var pe eventsv1.ProductEvent
	if err := proto.Unmarshal(msg.Value, &pe); err != nil {
		h.metrics.ParseErrors.Inc()
		log.Error("failed to unmarshal ProductEvent; skipping (poison pill)",
			zap.Error(err))
		return nil
	}

	event := &entities.ProductEvent{
		ID:         pe.GetId(),
		Name:       pe.GetName(),
		Type:       pe.GetType().String(),
		OccurredAt: pe.GetOccurredAt().AsTime(),
	}

	return h.service.Process(ctx, event)
}
