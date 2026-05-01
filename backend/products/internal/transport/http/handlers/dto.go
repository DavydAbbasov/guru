package handlers

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type ProductResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ListProductsResponse struct {
	Items []*ProductResponse `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}
