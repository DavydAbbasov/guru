package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const unmatchedRoute = "unmatched"

func (m *Metrics) FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		method := c.Method()

		m.HTTPInFlight.Inc()
		defer m.HTTPInFlight.Dec()

		err := c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Response().StatusCode()
		status := strconv.Itoa(statusCode)
		path := routeLabel(c, statusCode)

		m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}

func routeLabel(c *fiber.Ctx, status int) string {
	route := c.Route()
	if route == nil {
		return unmatchedRoute
	}
	path := route.Path
	if path == "" {
		return unmatchedRoute
	}
	if path == "/" && status == fiber.StatusNotFound {
		return unmatchedRoute
	}
	return path
}

func PrometheusHandler(m *Metrics) fiber.Handler {
	return adaptor.HTTPHandler(promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{
		EnableOpenMetrics: false,
		ErrorHandling:     promhttp.ContinueOnError,
	}))
}
