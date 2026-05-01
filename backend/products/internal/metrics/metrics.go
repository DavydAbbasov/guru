package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Created prometheus.Counter
	Deleted prometheus.Counter
	Active  prometheus.Gauge
}

func New(registry *prometheus.Registry, namespace string) *Metrics {
	m := &Metrics{
		Created: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "products",
			Name:      "created_total",
			Help:      "Total number of products created",
		}),
		Deleted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "products",
			Name:      "deleted_total",
			Help:      "Total number of products deleted",
		}),
		Active: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "products",
			Name:      "active",
			Help:      "Number of currently active products",
		}),
	}

	registry.MustRegister(
		m.Created,
		m.Deleted,
		m.Active,
	)

	return m
}
