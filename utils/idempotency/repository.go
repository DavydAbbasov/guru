package idempotency

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"guru/utils/logger"
)

type Repository struct {
	db      *gorm.DB
	metrics *Metrics
	log     logger.Logger
}

func NewRepository(db *gorm.DB, m *Metrics, log logger.Logger) *Repository {
	return &Repository{
		db:      db,
		metrics: m,
		log:     log.With(zap.String("component", "idempotency.repository")),
	}
}

func (r *Repository) MarkProcessed(ctx context.Context, id uuid.UUID, eventType string) (firstTime bool, err error) {
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ProcessedEvent{ID: id, EventType: eventType})
	if res.Error != nil {
		return false, fmt.Errorf("mark processed: %w", res.Error)
	}
	firstTime = res.RowsAffected == 1
	if firstTime {
		r.metrics.Processed.WithLabelValues(eventType).Inc()
	} else {
		r.metrics.Duplicate.WithLabelValues(eventType).Inc()
	}
	return firstTime, nil
}
