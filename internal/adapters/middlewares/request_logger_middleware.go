package middlewares

import (
	"time"

	"github.com/CSKU-Lab/main-server/internal/logging"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func RequestLoggerMiddleware(logger *zap.SugaredLogger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		requestID := uuid.New().String()

		c.Set("X-Request-ID", requestID)

		enriched := logger.With("request_id", requestID)
		sc := trace.SpanFromContext(c.Context()).SpanContext()
		if sc.IsValid() {
			enriched = enriched.With(
				"trace_id", sc.TraceID().String(),
				"span_id", sc.SpanID().String(),
			)
		}
		c.SetContext(logging.WithLogger(c.Context(), enriched))

		err := c.Next()

		enriched.Infow("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
		)

		return err
	}
}
