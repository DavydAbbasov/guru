package outbox

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	Pending          prometheus.Gauge
	Dispatched       *prometheus.CounterVec
	DispatchFailures *prometheus.CounterVec
	DispatchDuration prometheus.Histogram
	CleanupDeleted   prometheus.Counter
}

func NewMetrics(registry *prometheus.Registry, namespace string) *Metrics {
	m := &Metrics{
		Pending: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "outbox",
			Name:      "pending",
			Help:      "Outbox rows waiting to be published",
		}),
		Dispatched: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "outbox",
			Name:      "dispatched_total",
			Help:      "Outbox rows successfully published",
		}, []string{"event_type"}),
		DispatchFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "outbox",
			Name:      "dispatch_failures_total",
			Help:      "Outbox publish failures (will retry until MaxAttempts)",
		}, []string{"event_type"}),
		DispatchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "outbox",
			Name:      "dispatch_duration_seconds",
			Help:      "Time spent draining one batch",
			Buckets:   prometheus.DefBuckets,
		}),
		CleanupDeleted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "outbox",
			Name:      "cleanup_deleted_total",
			Help:      "Old sent rows removed by the cleanup tick",
		}),
	}
	registry.MustRegister(
		m.Pending,
		m.Dispatched,
		m.DispatchFailures,
		m.DispatchDuration,
		m.CleanupDeleted,
	)
	return m
}
