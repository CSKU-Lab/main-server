package middlewares

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"github.com/gofiber/fiber/v3"
)

// OtelMiddleware extracts W3C TraceContext from incoming request headers,
// starts a server span, and stores the context for downstream use.
func OtelMiddleware() fiber.Handler {
	propagator := otel.GetTextMapPropagator()
	tracer := otel.Tracer("main-server/http")

	return func(c fiber.Ctx) error {
		carrier := make(propagation.MapCarrier)
		c.Request().Header.VisitAll(func(k, v []byte) {
			carrier[string(k)] = string(v)
		})

		ctx := propagator.Extract(c.Context(), carrier)
		ctx, span := tracer.Start(ctx, c.Method()+" "+c.Route().Path)
		defer span.End()

		c.SetContext(ctx)
		return c.Next()
	}
}
