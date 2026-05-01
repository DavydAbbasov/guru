package idempotency

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	Processed      *prometheus.CounterVec
	Duplicate      *prometheus.CounterVec
	CleanupDeleted prometheus.Counter
}

func NewMetrics(registry *prometheus.Registry, namespace string) *Metrics {
	m := &Metrics{
		Processed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "idempotency",
			Name:      "processed_total",
			Help:      "Events processed for the first time",
		}, []string{"event_type"}),
		Duplicate: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "idempotency",
			Name:      "duplicate_total",
			Help:      "Duplicate events skipped",
		}, []string{"event_type"}),
		CleanupDeleted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "idempotency",
			Name:      "cleanup_deleted_total",
			Help:      "Old processed_events rows removed by cleanup tick",
		}),
	}
	registry.MustRegister(
		m.Processed,
		m.Duplicate,
		m.CleanupDeleted,
	)
	return m
}
