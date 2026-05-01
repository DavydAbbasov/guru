package idempotency

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"guru/utils/logger"
	"guru/utils/tracing"
)

type Cleanup struct {
	cfg     Config
	db      *gorm.DB
	tracer  *tracing.Tracer
	metrics *Metrics
	log     logger.Logger
}

func NewCleanup(cfg Config, db *gorm.DB, tracer *tracing.Tracer, m *Metrics, log logger.Logger) *Cleanup {
	cfg.withDefaults()
	return &Cleanup{
		cfg:     cfg,
		db:      db,
		tracer:  tracer,
		metrics: m,
		log:     log.With(zap.String("component", "idempotency.cleanup")),
	}
}

func (c *Cleanup) Run(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.CleanupInterval)
	defer ticker.Stop()

	c.log.Info("idempotency cleanup started",
		zap.Duration("cleanup_interval", c.cfg.CleanupInterval),
		zap.Duration("retention", c.cfg.Retention),
	)

	for {
		select {
		case <-ticker.C:
			c.cleanup(ctx)
		case <-ctx.Done():
			c.log.Info("idempotency cleanup stopped")
			return
		}
	}
}

func (c *Cleanup) cleanup(ctx context.Context) {
	ctx, span := c.tracer.Start(ctx, "idempotency.Cleanup.tick")
	defer span.End()

	threshold := time.Now().Add(-c.cfg.Retention)
	res := c.db.WithContext(ctx).
		Where("processed_at < ?", threshold).
		Delete(&ProcessedEvent{})
	if res.Error != nil {
		if !errors.Is(res.Error, context.Canceled) {
			c.log.Error("idempotency cleanup failed", zap.Error(res.Error))
		}
		return
	}
	if res.RowsAffected > 0 {
		c.metrics.CleanupDeleted.Add(float64(res.RowsAffected))
		c.log.Info("idempotency cleanup",
			zap.Int64("deleted", res.RowsAffected),
			zap.String("older_than", strconv.FormatFloat(c.cfg.Retention.Hours(), 'f', 1, 64)+"h"))
	}
}
