package repository

import "context"

// EventRepository tracks processed event IDs for idempotent consumption.
// MarkProcessed inserts the event_id atomically; returns true if this is the
// first time we see it, false if it's a duplicate (Kafka redelivery).
type EventRepository interface {
	MarkProcessed(ctx context.Context, eventID, eventType string) (firstTime bool, err error)
}
