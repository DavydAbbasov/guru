package outbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SaveParams struct {
	AggregateID uuid.UUID
	EventType   string
	Payload     []byte
}

type Builder struct{}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) Save(ctx context.Context, tx *gorm.DB, p SaveParams) error {
	if tx == nil {
		return fmt.Errorf("outbox.Save: tx is nil")
	}
	row := &Event{
		ID:          uuid.New(),
		AggregateID: p.AggregateID,
		EventType:   p.EventType,
		Payload:     p.Payload,
	}
	return tx.WithContext(ctx).Create(row).Error
}
