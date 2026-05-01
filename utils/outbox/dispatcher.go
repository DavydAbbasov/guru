package outbox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"guru/utils/logger"
	"guru/utils/tracing"
)

type Dispatcher struct {
	cfg       Config
	db        *gorm.DB
	publisher Publisher
	tracer    *tracing.Tracer
	metrics   *Metrics
	log       logger.Logger
}

func NewDispatcher(
	cfg Config,
	db *gorm.DB,
	publisher Publisher,
	tracer *tracing.Tracer,
	metrics *Metrics,
	log logger.Logger,
) *Dispatcher {
	cfg.withDefaults()
	return &Dispatcher{
		cfg:       cfg,
		db:        db,
		publisher: publisher,
		tracer:    tracer,
		metrics:   metrics,
		log:       log.With(zap.String("component", "outbox.dispatcher")),
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	dispatchTicker := time.NewTicker(d.cfg.PollInterval)
	cleanupTicker := time.NewTicker(d.cfg.CleanupInterval)
	defer dispatchTicker.Stop()
	defer cleanupTicker.Stop()

	d.log.Info("outbox dispatcher started",
		zap.Duration("poll_interval", d.cfg.PollInterval),
		zap.Duration("cleanup_interval", d.cfg.CleanupInterval),
		zap.Int("batch_size", d.cfg.BatchSize),
		zap.Int("max_attempts", d.cfg.MaxAttempts),
	)

	for {
		select {
		case <-dispatchTicker.C:
			d.dispatch(ctx)
		case <-cleanupTicker.C:
			d.cleanup(ctx)
		case <-ctx.Done():
			d.log.Info("outbox dispatcher stopped")
			return
		}
	}
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	ctx, span := d.tracer.Start(ctx, "outbox.Dispatcher.dispatch")
	defer span.End()

	start := time.Now()
	defer func() {
		d.metrics.DispatchDuration.Observe(time.Since(start).Seconds())
	}()

	d.refreshPending(ctx)

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var batch []Event
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("sent_at IS NULL AND next_retry_at <= NOW()").
			Order("next_retry_at").
			Limit(d.cfg.BatchSize).
			Find(&batch).Error; err != nil {
			return fmt.Errorf("fetch batch: %w", err)
		}
		for i := range batch {
			d.publishOne(ctx, tx, &batch[i])
		}
		return nil
	})
	if err != nil {
		d.log.Error("dispatch batch failed", zap.Error(err))
	}
}

func (d *Dispatcher) publishOne(ctx context.Context, tx *gorm.DB, ev *Event) {
	log := logger.FromContext(ctx, d.log).With(
		zap.String("event_id", ev.ID.String()),
		zap.String("event_type", ev.EventType),
		zap.Int("attempts", ev.Attempts),
	)
	headers := map[string]string{"outbox-event-id": ev.ID.String()}
	if err := d.publisher.Publish(ctx, ev.AggregateID.String(), ev.Payload, headers); err != nil {
		ev.Attempts++
		errStr := err.Error()
		ev.LastError = &errStr
		ev.NextRetryAt = time.Now().Add(d.backoff(ev.Attempts))
		d.metrics.DispatchFailures.WithLabelValues(ev.EventType).Inc()

		updates := map[string]any{
			"attempts":      ev.Attempts,
			"last_error":    errStr,
			"next_retry_at": ev.NextRetryAt,
		}
		if ev.Attempts >= d.cfg.MaxAttempts {
			now := time.Now()
			updates["sent_at"] = now // park: stop retrying; surface via metrics + log
			log.Error("outbox event abandoned after max attempts",
				zap.Error(err),
				zap.Int("max_attempts", d.cfg.MaxAttempts))
		} else {
			log.Warn("outbox publish failed; will retry",
				zap.Error(err),
				zap.Time("next_retry_at", ev.NextRetryAt))
		}
		if e := tx.Model(&Event{}).Where("id = ?", ev.ID).Updates(updates).Error; e != nil {
			log.Error("failed to record publish failure", zap.Error(e))
		}
		return
	}

	now := time.Now()
	if err := tx.Model(&Event{}).Where("id = ?", ev.ID).
		Updates(map[string]any{"sent_at": now, "last_error": nil}).Error; err != nil {
		log.Error("failed to mark sent", zap.Error(err))
		return
	}
	d.metrics.Dispatched.WithLabelValues(ev.EventType).Inc()
}

func (d *Dispatcher) cleanup(ctx context.Context) {
	threshold := time.Now().Add(-d.cfg.Retention)
	res := d.db.WithContext(ctx).
		Where("sent_at IS NOT NULL AND sent_at < ?", threshold).
		Delete(&Event{})
	if res.Error != nil {
		if !errors.Is(res.Error, context.Canceled) {
			d.log.Error("outbox cleanup failed", zap.Error(res.Error))
		}
		return
	}
	if res.RowsAffected > 0 {
		d.metrics.CleanupDeleted.Add(float64(res.RowsAffected))
		d.log.Info("outbox cleanup",
			zap.Int64("deleted", res.RowsAffected),
			zap.String("older_than", strconv.FormatFloat(d.cfg.Retention.Hours(), 'f', 1, 64)+"h"))
	}
}

func (d *Dispatcher) refreshPending(ctx context.Context) {
	var n int64
	if err := d.db.WithContext(ctx).Model(&Event{}).
		Where("sent_at IS NULL").
		Count(&n).Error; err != nil {
		return
	}
	d.metrics.Pending.Set(float64(n))
}

func (d *Dispatcher) backoff(attempts int) time.Duration {
	d2 := d.cfg.RetryBaseDelay
	for i := 1; i < attempts; i++ {
		d2 *= 2
		if d2 > time.Hour {
			return time.Hour
		}
	}
	return d2
}
