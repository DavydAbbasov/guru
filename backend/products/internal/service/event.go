package service

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventsv1 "guru/apis/products/v1/events"
	"guru/backend/products/internal/entities"
)

func buildProductEventPayload(product *entities.Product, eventType eventsv1.EventType) ([]byte, error) {
	if product == nil {
		return nil, fmt.Errorf("nil product")
	}
	event := &eventsv1.ProductEvent{
		Id:         product.ID.String(),
		Name:       product.Name,
		Type:       eventType,
		OccurredAt: timestamppb.New(time.Now()),
	}
	payload, err := proto.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal product event: %w", err)
	}
	return payload, nil
}
