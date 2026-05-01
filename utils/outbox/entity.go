package outbox

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID  `gorm:"primaryKey;column:id"`
	AggregateID uuid.UUID  `gorm:"column:aggregate_id;not null"`
	EventType   string     `gorm:"column:event_type;not null"`
	Payload     []byte     `gorm:"column:payload;not null"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:now()"`
	SentAt      *time.Time `gorm:"column:sent_at"`
	Attempts    int        `gorm:"column:attempts;not null;default:0"`
	NextRetryAt time.Time  `gorm:"column:next_retry_at;not null;default:now()"`
	LastError   *string    `gorm:"column:last_error"`
}

func (Event) TableName() string { return "outbox" }
