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
	"gorm.io/gorm"

	eventsv1 "guru/apis/products/v1/events"
	"guru/backend/products/internal/entities"
	productsmetrics "guru/backend/products/internal/metrics"
	"guru/backend/products/internal/repository"
	customerrors "guru/utils/custom-errors"
	"guru/utils/logger"
	"guru/utils/outbox"
	"guru/utils/tracing"
)

type fakeRepo struct {
	createFn func(*entities.Product) error
	deleteFn func(uuid.UUID) (*entities.Product, error)
	listFn   func(limit, offset int) ([]entities.Product, int64, error)

	created    []*entities.Product
	deletedIDs []uuid.UUID
	listCalls  []listCall
}

type listCall struct{ limit, offset int }

func (r *fakeRepo) Create(_ context.Context, _ *gorm.DB, p *entities.Product) error {
	r.created = append(r.created, p)
	if r.createFn != nil {
		return r.createFn(p)
	}
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, _ *gorm.DB, id uuid.UUID) (*entities.Product, error) {
	r.deletedIDs = append(r.deletedIDs, id)
	if r.deleteFn != nil {
		return r.deleteFn(id)
	}
	return &entities.Product{ID: id, Name: "deleted"}, nil
}

func (r *fakeRepo) List(_ context.Context, limit, offset int) ([]entities.Product, int64, error) {
	r.listCalls = append(r.listCalls, listCall{limit, offset})
	if r.listFn != nil {
		return r.listFn(limit, offset)
	}
	return nil, 0, nil
}

func (r *fakeRepo) FindByID(_ context.Context, _ uuid.UUID) (*entities.Product, error) {
	return nil, repository.ErrNotFound
}

type fakeOutbox struct {
	failWith error
	saved    []outbox.SaveParams
}

func (o *fakeOutbox) Save(_ context.Context, _ *gorm.DB, p outbox.SaveParams) error {
	o.saved = append(o.saved, p)
	return o.failWith
}

type noopLogger struct{}

func (noopLogger) Info(string, ...zap.Field)       {}
func (noopLogger) Debug(string, ...zap.Field)      {}
func (noopLogger) Warn(string, ...zap.Field)       {}
func (noopLogger) Error(string, ...zap.Field)      {}
func (noopLogger) Fatal(string, ...zap.Field)      {}
func (noopLogger) With(...zap.Field) logger.Logger { return noopLogger{} }
func (noopLogger) Sync() error                     { return nil }

type passthroughTxMgr struct{}

func (passthroughTxMgr) WithTransaction(_ context.Context, fn func(tx *gorm.DB) error) error {
	return fn(nil)
}

func newTestService(t *testing.T) (*ProductService, *fakeRepo, *fakeOutbox, *productsmetrics.Metrics) {
	t.Helper()

	repo := &fakeRepo{}
	ob := &fakeOutbox{}
	tracer, err := tracing.New(&tracing.Config{Disabled: true})
	require.NoError(t, err)
	m := productsmetrics.New(prometheus.NewRegistry(), "test")

	svc := NewProductService(repo, ob, passthroughTxMgr{}, tracer, m, noopLogger{})
	return svc, repo, ob, m
}

func counterValue(t *testing.T, c prometheus.Counter) float64 {
	t.Helper()
	m := &dto.Metric{}
	require.NoError(t, c.Write(m))
	return m.GetCounter().GetValue()
}

func gaugeValue(t *testing.T, g prometheus.Gauge) float64 {
	t.Helper()
	m := &dto.Metric{}
	require.NoError(t, g.Write(m))
	return m.GetGauge().GetValue()
}

func TestProductService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the persisted product on success", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)

		got, err := svc.Create(ctx, "yoga mat")

		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "yoga mat", got.Name)
		require.NotEqual(t, uuid.Nil, got.ID)
		require.False(t, got.CreatedAt.IsZero())
		require.Len(t, repo.created, 1, "repo.Create should be called exactly once")
		require.Equal(t, got.ID, repo.created[0].ID)
	})

	t.Run("writes a CREATED event into the outbox in the same tx", func(t *testing.T) {
		svc, _, ob, _ := newTestService(t)

		got, err := svc.Create(ctx, "candle")
		require.NoError(t, err)

		require.Len(t, ob.saved, 1)
		require.Equal(t, eventsv1.EventType_EVENT_TYPE_CREATED.String(), ob.saved[0].EventType)
		require.Equal(t, got.ID, ob.saved[0].AggregateID)
		require.NotEmpty(t, ob.saved[0].Payload, "outbox payload must contain marshaled protobuf")
	})

	t.Run("increments the created and active counters", func(t *testing.T) {
		svc, _, _, m := newTestService(t)

		_, err := svc.Create(ctx, "phone case")
		require.NoError(t, err)

		require.Equal(t, 1.0, counterValue(t, m.Created))
		require.Equal(t, 1.0, gaugeValue(t, m.Active))
	})

	t.Run("rejects empty or whitespace-only names with ErrValidation", func(t *testing.T) {
		svc, repo, ob, _ := newTestService(t)

		for _, name := range []string{"", "   ", "\t\n"} {
			t.Run("name="+name, func(t *testing.T) {
				_, err := svc.Create(ctx, name)
				require.ErrorIs(t, err, customerrors.ErrValidation)
			})
		}

		require.Empty(t, repo.created, "repo must not be touched on validation error")
		require.Empty(t, ob.saved, "no event must be written on validation error")
	})

	t.Run("hides infrastructure errors behind ErrInternal", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)
		repo.createFn = func(*entities.Product) error { return errors.New("connection refused") }

		_, err := svc.Create(ctx, "lamp")

		require.ErrorIs(t, err, customerrors.ErrInternal)
		require.NotContains(t, err.Error(), "connection refused", "raw repo error must not leak to caller")
	})

	t.Run("rolls back the request when outbox.Save fails", func(t *testing.T) {
		svc, _, ob, m := newTestService(t)
		ob.failWith = errors.New("outbox unavailable")

		_, err := svc.Create(ctx, "umbrella")

		require.ErrorIs(t, err, customerrors.ErrInternal,
			"if outbox cannot record the event the whole tx must fail; otherwise we'd have a product without an event")
		require.Equal(t, 0.0, counterValue(t, m.Created),
			"metrics must reflect the rolled-back state")
	})
}

func TestProductService_Delete(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	t.Run("removes the product and writes a DELETED outbox event", func(t *testing.T) {
		svc, repo, ob, _ := newTestService(t)

		err := svc.Delete(ctx, id)

		require.NoError(t, err)
		require.Equal(t, []uuid.UUID{id}, repo.deletedIDs)
		require.Len(t, ob.saved, 1)
		require.Equal(t, eventsv1.EventType_EVENT_TYPE_DELETED.String(), ob.saved[0].EventType)
		require.Equal(t, id, ob.saved[0].AggregateID)
	})

	t.Run("increments deleted and decrements active", func(t *testing.T) {
		svc, _, _, m := newTestService(t)

		require.NoError(t, svc.Delete(ctx, id))

		require.Equal(t, 1.0, counterValue(t, m.Deleted))
		require.Equal(t, -1.0, gaugeValue(t, m.Active), "Active is signed; Delete only decrements")
	})

	t.Run("translates repository.ErrNotFound to ErrNotFound", func(t *testing.T) {
		svc, repo, ob, _ := newTestService(t)
		repo.deleteFn = func(uuid.UUID) (*entities.Product, error) {
			return nil, repository.ErrNotFound
		}

		err := svc.Delete(ctx, id)

		require.ErrorIs(t, err, customerrors.ErrNotFound)
		require.Empty(t, ob.saved, "no event must be written when the product was not found")
	})

	t.Run("hides arbitrary repo errors behind ErrInternal", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)
		repo.deleteFn = func(uuid.UUID) (*entities.Product, error) {
			return nil, errors.New("deadlock detected")
		}

		err := svc.Delete(ctx, id)

		require.ErrorIs(t, err, customerrors.ErrInternal)
	})
}

func TestProductService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("forwards the page/limit pair as a SQL offset", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)

		_, _, err := svc.List(ctx, 3, 25)

		require.NoError(t, err)
		require.Equal(t, []listCall{{limit: 25, offset: 50}}, repo.listCalls,
			"page=3, limit=25 → offset=(3-1)*25=50")
	})

	t.Run("clamps non-positive page to 1", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)

		_, _, err := svc.List(ctx, 0, 20)

		require.NoError(t, err)
		require.Equal(t, 0, repo.listCalls[0].offset, "page=0 should behave as page=1 → offset=0")
	})

	t.Run("falls back to the default limit when out of range", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)

		_, _, err := svc.List(ctx, 1, 9999)
		require.NoError(t, err)
		require.Equal(t, 10, repo.listCalls[0].limit, "limit > 100 must clamp to the default 10")

		repo.listCalls = nil
		_, _, err = svc.List(ctx, 1, 0)
		require.NoError(t, err)
		require.Equal(t, 10, repo.listCalls[0].limit, "limit < 1 must clamp to the default 10")
	})

	t.Run("returns the products and total count from the repository", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)
		want := []entities.Product{{ID: uuid.New(), Name: "a"}, {ID: uuid.New(), Name: "b"}}
		repo.listFn = func(int, int) ([]entities.Product, int64, error) { return want, 42, nil }

		got, total, err := svc.List(ctx, 1, 20)

		require.NoError(t, err)
		require.Equal(t, want, got)
		require.Equal(t, int64(42), total)
	})

	t.Run("hides repository errors behind ErrInternal", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)
		repo.listFn = func(int, int) ([]entities.Product, int64, error) {
			return nil, 0, errors.New("connection lost")
		}

		_, _, err := svc.List(ctx, 1, 20)

		require.ErrorIs(t, err, customerrors.ErrInternal)
	})
}
