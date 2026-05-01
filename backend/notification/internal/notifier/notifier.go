package notifier

import (
	"context"

	"guru/backend/notification/internal/entities"
)

type Notifier interface {
	Notify(ctx context.Context, event *entities.ProductEvent) error
}
