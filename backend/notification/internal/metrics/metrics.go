package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	EventsConsumed     *prometheus.CounterVec
	EventsProcessed    *prometheus.CounterVec
	EventsFailed       *prometheus.CounterVec
	ParseErrors        prometheus.Counter
	ProcessingDuration *prometheus.HistogramVec
}

func New(registry *prometheus.Registry, namespace string) *Metrics {
	m := &Metrics{
		EventsConsumed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "events",
			Name:      "consumed_total",
			Help:      "Total events received from Kafka",
		}, []string{"type"}),
		EventsProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "events",
			Name:      "processed_total",
			Help:      "Total events successfully processed",
		}, []string{"type"}),
		EventsFailed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "events",
			Name:      "failed_total",
			Help:      "Total events that failed processing",
		}, []string{"type", "reason"}),
		ParseErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "events",
			Name:      "parse_errors_total",
			Help:      "Total messages that failed protobuf decoding (poison pills)",
		}),
		ProcessingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "events",
			Name:      "processing_duration_seconds",
			Help:      "Time spent processing an event end-to-end",
			Buckets:   prometheus.DefBuckets,
		}, []string{"type"}),
	}

	registry.MustRegister(
		m.EventsConsumed,
		m.EventsProcessed,
		m.EventsFailed,
		m.ParseErrors,
		m.ProcessingDuration,
	)

	return m
}
