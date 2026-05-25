package middleware

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"market-core/internal/platform/logger"
	"market-core/internal/platform/metrics"
)

func Logger(log *slog.Logger, m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// inject trace_id from the active OTel span so every log line inside handlers carries it
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			ctx := logger.WithTraceID(c.Request.Context(), traceID)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		log.Info("request",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
		)

		if m != nil {
			m.HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
			m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
		}
	}
}
