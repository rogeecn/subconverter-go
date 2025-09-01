package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subconverter_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "subconverter_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "subconverter_active_connections",
			Help: "Number of active connections",
		},
	)

	conversionRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subconverter_conversion_requests_total",
			Help: "Total number of conversion requests",
		},
		[]string{"target", "status"},
	)

	conversionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "subconverter_conversion_duration_seconds",
			Help:    "Conversion duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"target"},
	)

	cacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "subconverter_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "subconverter_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)

// MetricsMiddleware adds Prometheus metrics to HTTP requests
func MetricsMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		
		// Increment active connections
		activeConnections.Inc()
		defer activeConnections.Dec()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		
		httpRequestsTotal.WithLabelValues(
			c.Method(),
			c.Path(),
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Method(),
			c.Path(),
			status,
		).Observe(duration)

		return err
	}
}

// ConversionMetrics records conversion-specific metrics
func ConversionMetrics(target string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	conversionRequestsTotal.WithLabelValues(target, status).Inc()
	conversionDuration.WithLabelValues(target).Observe(duration.Seconds())
}

// CacheMetrics records cache hit/miss metrics
func CacheMetrics(hit bool) {
	if hit {
		cacheHits.Inc()
	} else {
		cacheMisses.Inc()
	}
}