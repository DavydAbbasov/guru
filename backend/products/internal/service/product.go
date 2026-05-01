package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	eventsv1 "guru/apis/products/v1/events"
	"guru/backend/products/internal/entities"
	productsmetrics "guru/backend/products/internal/metrics"
	"guru/backend/products/internal/publisher"
	"guru/backend/products/internal/repository"
	customerrors "guru/utils/custom-errors"
	"guru/utils/logger"
	"guru/utils/tracing"
)

type ProductService struct {
	repo      repository.ProductRepository
	publisher publisher.EventPublisher
	tracer    *tracing.Tracer
	metrics   *productsmetrics.Metrics
	log       logger.Logger
}

func NewProductService(
	repo repository.ProductRepository,
	pub publisher.EventPublisher,
	tracer *tracing.Tracer,
	m *productsmetrics.Metrics,
	log logger.Logger,
) *ProductService {
	return &ProductService{
		repo:      repo,
		publisher: pub,
		tracer:    tracer,
		metrics:   m,
		log:       log,
	}
}

func (s *ProductService) Create(ctx context.Context, name string) (*entities.Product, error) {
	ctx, span := s.tracer.Start(ctx, "ProductService.Create")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required: %w", customerrors.ErrValidation)
	}

	product := &entities.Product{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, s.mapRepoError(ctx, "failed to create product", err)
	}

	s.metrics.Created.Inc()
	s.metrics.Active.Inc()

	if err := s.publisher.PublishProductEvent(
		ctx,
		product,
		eventsv1.EventType_EVENT_TYPE_CREATED); err != nil {
		log.Error("failed to publish product created event",
			zap.Error(err))
	}

	log.Info("product created",
		zap.String("id", product.ID.String()),
		zap.String("name", product.Name))
	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "ProductService.Delete")
	defer span.End()

	log := logger.FromContext(ctx, s.log)

	product, err := s.repo.Delete(ctx, id)
	if err != nil {
		return s.mapRepoError(ctx, "failed to delete product", err)
	}

	s.metrics.Deleted.Inc()
	s.metrics.Active.Dec()

	if err := s.publisher.PublishProductEvent(
		ctx,
		product,
		eventsv1.EventType_EVENT_TYPE_DELETED); err != nil {
		log.Error("failed to publish product deleted event",
			zap.Error(err))
	}

	log.Info("product deleted",
		zap.String("id", id.String()))
	return nil
}

func (s *ProductService) List(ctx context.Context, page, limit int) ([]entities.Product, int64, error) {
	ctx, span := s.tracer.Start(ctx, "ProductService.List")
	defer span.End()

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	products, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, s.mapRepoError(ctx, "failed to list products", err)
	}

	return products, total, nil
}

func (s *ProductService) mapRepoError(ctx context.Context, msg string, err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return customerrors.ErrNotFound
	}
	logger.FromContext(ctx, s.log).Error(msg, zap.Error(err))
	return customerrors.ErrInternal
}
