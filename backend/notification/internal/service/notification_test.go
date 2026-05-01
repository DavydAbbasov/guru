package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"guru/backend/notification/internal/entities"
	notificationmetrics "guru/backend/notification/internal/metrics"
	"guru/utils/logger"
	"guru/utils/tracing"
)

type fakeNotifier struct {
	failWith error
	notified []*entities.ProductEvent
}

func (n *fakeNotifier) Notify(_ context.Context, e *entities.ProductEvent) error {
	n.notified = append(n.notified, e)
	return n.failWith
}

type fakeIdem struct {
	seen     map[uuid.UUID]bool
	failWith error
	calls    int
}

func newFakeIdem() *fakeIdem { return &fakeIdem{seen: map[uuid.UUID]bool{}} }

func (f *fakeIdem) MarkProcessed(_ context.Context, id uuid.UUID, _ string) (bool, error) {
	f.calls++
	if f.failWith != nil {
		return false, f.failWith
	}
	if f.seen[id] {
		return false, nil
	}
	f.seen[id] = true
	return true, nil
}

type noopLogger struct{}

func (noopLogger) Info(string, ...zap.Field)       {}
func (noopLogger) Debug(string, ...zap.Field)      {}
func (noopLogger) Warn(string, ...zap.Field)       {}
func (noopLogger) Error(string, ...zap.Field)      {}
func (noopLogger) Fatal(string, ...zap.Field)      {}
func (noopLogger) With(...zap.Field) logger.Logger { return noopLogger{} }
func (noopLogger) Sync() error                     { return nil }

func newTestService(t *testing.T) (*NotificationService, *fakeNotifier, *fakeIdem, *notificationmetrics.Metrics) {
	t.Helper()

	n := &fakeNotifier{}
	idem := newFakeIdem()
	tracer, err := tracing.New(&tracing.Config{Disabled: true})
	require.NoError(t, err)
	m := notificationmetrics.New(prometheus.NewRegistry(), "test")

	svc := NewNotificationService(n, idem, m, tracer, noopLogger{})
	return svc, n, idem, m
}

func counterValueLabels(t *testing.T, c *prometheus.CounterVec, labels ...string) float64 {
	t.Helper()
	m := &dto.Metric{}
	require.NoError(t, c.WithLabelValues(labels...).Write(m))
	return m.GetCounter().GetValue()
}

func histogramSampleCount(t *testing.T, h *prometheus.HistogramVec, labels ...string) uint64 {
	t.Helper()
	m := &dto.Metric{}
	require.NoError(t, h.WithLabelValues(labels...).(prometheus.Metric).Write(m))
	return m.GetHistogram().GetSampleCount()
}

func validEvent() *entities.ProductEvent {
	return &entities.ProductEvent{
		ID:   "evt-1",
		Name: "yoga mat",
		Type: "EVENT_TYPE_CREATED",
	}
}

func newEventID() string { return uuid.NewString() }

func TestNotificationService_Process(t *testing.T) {
	ctx := context.Background()

	t.Run("dispatches the event to the notifier on success", func(t *testing.T) {
		svc, n, _, _ := newTestService(t)
		evt := validEvent()

		err := svc.Process(ctx, evt, newEventID())

		require.NoError(t, err)
		require.Len(t, n.notified, 1)
		require.Same(t, evt, n.notified[0], "notifier must receive the original event by reference")
	})

	t.Run("counts a successful event as both consumed and processed", func(t *testing.T) {
		svc, _, _, m := newTestService(t)

		require.NoError(t, svc.Process(ctx, validEvent(), newEventID()))

		require.Equal(t, 1.0, counterValueLabels(t, m.EventsConsumed, "EVENT_TYPE_CREATED"))
		require.Equal(t, 1.0, counterValueLabels(t, m.EventsProcessed, "EVENT_TYPE_CREATED"))
	})

	t.Run("always observes processing duration, even on failure", func(t *testing.T) {
		svc, n, _, m := newTestService(t)
		n.failWith = errors.New("smtp down")

		_ = svc.Process(ctx, validEvent(), newEventID())

		require.Equal(t, uint64(1), histogramSampleCount(t, m.ProcessingDuration, "EVENT_TYPE_CREATED"))
	})

	t.Run("uses 'unknown' label when event type is empty", func(t *testing.T) {
		svc, _, _, m := newTestService(t)
		evt := validEvent()
		evt.Type = ""

		_ = svc.Process(ctx, evt, newEventID())

		require.Equal(t, 1.0, counterValueLabels(t, m.EventsConsumed, "unknown"))
		require.Equal(t, 1.0, counterValueLabels(t, m.EventsFailed, "unknown", "validation"),
			"empty Type also fails validation, so failure is bucketed under 'unknown'")
	})
}

func TestNotificationService_Process_dedup(t *testing.T) {
	ctx := context.Background()

	t.Run("processes a fresh event once and records it as processed", func(t *testing.T) {
		svc, n, idem, m := newTestService(t)
		eventID := newEventID()

		require.NoError(t, svc.Process(ctx, validEvent(), eventID))

		require.Equal(t, 1, idem.calls, "MarkProcessed must be called once for a fresh event")
		require.Len(t, n.notified, 1, "notifier must run for a first-time event")
		require.Equal(t, 1.0, counterValueLabels(t, m.EventsProcessed, "EVENT_TYPE_CREATED"))
	})

	t.Run("skips duplicates: notifier is not called again on the same outbox-event-id", func(t *testing.T) {
		svc, n, _, m := newTestService(t)
		eventID := newEventID()

		require.NoError(t, svc.Process(ctx, validEvent(), eventID))
		require.NoError(t, svc.Process(ctx, validEvent(), eventID))

		require.Len(t, n.notified, 1, "notifier must run only on the first delivery; the second is a duplicate")
		require.Equal(t, 1.0, counterValueLabels(t, m.EventsProcessed, "EVENT_TYPE_CREATED"),
			"only the first-time event counts as processed; duplicates do not bump the counter")
	})

	t.Run("processes the event when no outbox-event-id header is present (dedup is best-effort)", func(t *testing.T) {
		svc, n, idem, _ := newTestService(t)

		require.NoError(t, svc.Process(ctx, validEvent(), ""))

		require.Equal(t, 0, idem.calls, "no header → MarkProcessed must not be called")
		require.Len(t, n.notified, 1, "notifier must still run when the dedup id is missing")
	})

	t.Run("propagates a dedup repository error to the caller", func(t *testing.T) {
		svc, n, idem, m := newTestService(t)
		idem.failWith = errors.New("db down")

		err := svc.Process(ctx, validEvent(), newEventID())

		require.Error(t, err)
		require.EqualError(t, err, "db down")
		require.Empty(t, n.notified, "notifier must not run when dedup fails")
		require.Equal(t, 1.0, counterValueLabels(t, m.EventsFailed, "EVENT_TYPE_CREATED", "dedup"))
	})

	t.Run("malformed outbox-event-id is bucketed and falls through to processing (best-effort dedup)", func(t *testing.T) {
		svc, n, idem, m := newTestService(t)

		require.NoError(t, svc.Process(ctx, validEvent(), "not-a-uuid"))

		require.Equal(t, 0, idem.calls, "MarkProcessed must not be called with an unparseable id")
		require.Len(t, n.notified, 1, "service should still notify rather than drop the event")
		require.Equal(t, 1.0, counterValueLabels(t, m.EventsFailed, "EVENT_TYPE_CREATED", "bad_event_id"))
	})
}

func TestNotificationService_Process_validation(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		evt  *entities.ProductEvent
	}{
		{"nil event", nil},
		{"empty id", &entities.ProductEvent{ID: "", Type: "EVENT_TYPE_CREATED"}},
		{"whitespace id", &entities.ProductEvent{ID: "   ", Type: "EVENT_TYPE_CREATED"}},
		{"empty type", &entities.ProductEvent{ID: "evt-1", Type: ""}},
		{"whitespace type", &entities.ProductEvent{ID: "evt-1", Type: "\t"}},
	}

	for _, tc := range cases {
		t.Run(tc.name+" is rejected with ErrInvalidEvent", func(t *testing.T) {
			svc, n, _, m := newTestService(t)

			err := svc.Process(ctx, tc.evt, newEventID())

			require.ErrorIs(t, err, ErrInvalidEvent)
			require.Empty(t, n.notified, "notifier must not be called when validation fails")
			require.Equal(t, 0.0, counterValueLabels(t, m.EventsProcessed, "EVENT_TYPE_CREATED"),
				"failed validation must not count as processed")
		})
	}
}

func TestNotificationService_Process_notifierFailure(t *testing.T) {
	ctx := context.Background()

	t.Run("propagates notifier error to the caller", func(t *testing.T) {
		svc, n, _, _ := newTestService(t)
		n.failWith = errors.New("smtp down")

		err := svc.Process(ctx, validEvent(), newEventID())

		require.Error(t, err)
		require.EqualError(t, err, "smtp down")
	})

	t.Run("buckets the failure under reason='notify'", func(t *testing.T) {
		svc, n, _, m := newTestService(t)
		n.failWith = errors.New("smtp down")

		_ = svc.Process(ctx, validEvent(), newEventID())

		require.Equal(t, 1.0, counterValueLabels(t, m.EventsFailed, "EVENT_TYPE_CREATED", "notify"))
		require.Equal(t, 0.0, counterValueLabels(t, m.EventsProcessed, "EVENT_TYPE_CREATED"),
			"a failed notify must not count as processed")
	})
}
