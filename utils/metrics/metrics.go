package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Registry *prometheus.Registry

	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPInFlight        prometheus.Gauge

	ErrorsTotal *prometheus.CounterVec

	DBQueryDuration *prometheus.HistogramVec
}

func New(namespace string) *Metrics {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		Registry: reg,
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"method", "path", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		}, []string{"method", "path"}),
		HTTPInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		}),
		ErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Total number of errors",
		}, []string{"type"}),
		DBQueryDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}, []string{"operation", "table"}),
	}

	reg.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPInFlight,
		m.ErrorsTotal,
		m.DBQueryDuration,
	)

	return m
}
