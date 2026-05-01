package entities

import "time"

type ProductEvent struct {
	ID         string
	Name       string
	Type       string
	OccurredAt time.Time
}
