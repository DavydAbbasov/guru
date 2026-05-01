package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID        uuid.UUID      `json:"id"         gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `json:"name"       gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"          gorm:"index"`
}

func (Product) TableName() string {
	return "products"
}
