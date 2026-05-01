package repository

import (
	"context"
	"errors"

	"guru/backend/products/internal/entities"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("product not found")

type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	List(ctx context.Context, limit, offset int) ([]entities.Product, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
}
