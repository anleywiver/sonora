package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Sprint 13 (ADR 0007): deliberately minimal — request count + latency by
// method/route/status, plus whatever Go runtime metrics client_golang
// registers by default. No business-specific metrics (queue depth, etc.)
// — personal scale just needs "is the API up and not slow".
var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sonora_http_requests_total",
		Help: "Total HTTP requests processed by the API",
	}, []string{"method", "route", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "sonora_http_request_duration_seconds",
		Help: "HTTP request duration in seconds",
	}, []string{"method", "route"})
)

// Metrics records request count/duration for every request. Registered
// before routes so c.Route().Path reflects the matched route pattern
// (e.g. "/songs/:id"), not the raw URL.
func Metrics(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()

	route := c.Route().Path
	status := strconv.Itoa(c.Response().StatusCode())
	httpRequestsTotal.WithLabelValues(c.Method(), route, status).Inc()
	httpRequestDuration.WithLabelValues(c.Method(), route).Observe(time.Since(start).Seconds())

	return err
}
