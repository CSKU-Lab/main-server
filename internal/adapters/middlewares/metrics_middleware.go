package middlewares

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_request_duration_seconds",
		Help:    "Latency of HTTP requests by method, route template and status class.",
		Buckets: []float64{0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method", "route", "status"})

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_server_requests_in_flight",
		Help: "Number of HTTP requests currently being served.",
	})
)

// MetricsMiddleware records request latency and in-flight requests. It must be
// registered first so the measured latency covers every other middleware.
//
// Errors are resolved through the app error handler here, like Fiber's logger
// middleware does, so the recorded status is the one the client receives.
// Labels use the route template (e.g. /api/v1/labs/:id), never the raw path,
// and the status class (2xx, 4xx, ...) instead of the exact code, to keep
// series count within the Grafana Cloud free tier.
func MetricsMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		if err := c.Next(); err != nil {
			if hErr := c.App().ErrorHandler(c, err); hErr != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		httpRequestDuration.WithLabelValues(
			c.Method(),
			c.Route().Path,
			statusClass(c.Response().StatusCode()),
		).Observe(time.Since(start).Seconds())

		return nil
	}
}

func statusClass(code int) string {
	return strconv.Itoa(code/100) + "xx"
}
