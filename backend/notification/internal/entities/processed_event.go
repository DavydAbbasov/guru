package entities

import "time"

type ProcessedEvent struct {
	EventID     string    `gorm:"primaryKey;column:event_id"`
	EventType   string    `gorm:"column:event_type;not null"`
	ProcessedAt time.Time `gorm:"column:processed_at;not null;default:now()"`
}

func (ProcessedEvent) TableName() string {
	return "processed_events"
}
