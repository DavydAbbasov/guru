package idempotency

import (
	"time"

	"github.com/google/uuid"
)

type ProcessedEvent struct {
	ID          uuid.UUID `gorm:"primaryKey;column:id"`
	EventType   string    `gorm:"column:event_type;not null"`
	ProcessedAt time.Time `gorm:"column:processed_at;not null;default:now()"`
}

func (ProcessedEvent) TableName() string { return "processed_events" }
