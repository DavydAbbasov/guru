package pgsql

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"guru/backend/notification/internal/entities"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) MarkProcessed(ctx context.Context, eventID, eventType string) (bool, error) {
	row := &entities.ProcessedEvent{EventID: eventID, EventType: eventType}
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(row)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}
