package notifier

import (
	"context"

	"go.uber.org/zap"

	"guru/backend/notification/internal/entities"
	"guru/utils/logger"
)

type LogNotifier struct {
	log logger.Logger
}

func NewLogNotifier(log logger.Logger) *LogNotifier {
	return &LogNotifier{log: log.With(zap.String("component", "notifier.log"))}
}

func (n *LogNotifier) Notify(ctx context.Context, event *entities.ProductEvent) error {
	logger.FromContext(ctx, n.log).Info("notification dispatched",
		zap.String("event_id", event.ID),
		zap.String("type", event.Type),
		zap.String("name", event.Name),
	)
	return nil
}
