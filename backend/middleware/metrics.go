package middleware

import (
	"chemistryuno/backend/metrics"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware tracks API request duration and records to Prometheus
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics after request completes
		duration := time.Since(start).Seconds()
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		metrics.APIRequestDuration.WithLabelValues(endpoint, method, status).Observe(duration)
	}
}
