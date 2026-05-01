package publisher

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventsv1 "guru/apis/products/v1/events"
	"guru/backend/products/internal/entities"
	kafkatool "guru/utils/kafka-tool"
	"guru/utils/tracing"
)

type KafkaPublisher struct {
	producer *kafkatool.Producer
	tracer   *tracing.Tracer
}

func NewKafkaPublisher(producer *kafkatool.Producer, tracer *tracing.Tracer) *KafkaPublisher {
	return &KafkaPublisher{
		producer: producer,
		tracer:   tracer,
	}
}

func (p *KafkaPublisher) PublishProductEvent(
	ctx context.Context,
	product *entities.Product,
	eventType eventsv1.EventType,
) error {
	ctx, span := p.tracer.Start(ctx, "KafkaPublisher.PublishProductEvent",
		trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	event := &eventsv1.ProductEvent{
		Id:         product.ID.String(),
		Name:       product.Name,
		Type:       eventType,
		OccurredAt: timestamppb.New(time.Now()),
	}

	payload, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	partition, offset, err := p.producer.SendMessage(ctx, eventType.String(), payload)
	if err != nil {
		return fmt.Errorf("failed to send kafka message: %w", err)
	}

	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination", p.producer.Topic()),
		attribute.String("messaging.kafka.message_key", eventType.String()),
		attribute.Int64("messaging.kafka.partition", int64(partition)),
		attribute.Int64("messaging.kafka.offset", offset),
	)

	return nil
}
