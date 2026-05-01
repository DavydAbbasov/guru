package pgsql

import (
	"context"
	"errors"

	"guru/backend/products/internal/entities"
	"guru/backend/products/internal/repository"

	"github.com/google/uuid"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *entities.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", id).
				First(&product).Error; err != nil {
				return err
			}
			return tx.Delete(&product).Error
		})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]entities.Product, int64, error) {
	var products []entities.Product
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&entities.Product{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.WithContext(ctx).
		Order("created_at DESC, id DESC").
		Offset(offset).
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	var product entities.Product
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}
