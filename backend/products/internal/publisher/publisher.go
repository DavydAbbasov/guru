package publisher

import (
	"context"

	eventsv1 "guru/apis/products/v1/events"
	"guru/backend/products/internal/entities"
)

type EventPublisher interface {
	PublishProductEvent(ctx context.Context, product *entities.Product, eventType eventsv1.EventType) error
}
