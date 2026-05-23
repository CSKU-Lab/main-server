package middlewares

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/adapters/ratelimit"
	"github.com/gofiber/fiber/v3"
)

func RateLimitMiddleware(rl ratelimit.RateLimiter, limit int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		key := fmt.Sprintf("rate_limit:core:%s", user.Username)

		allowed, retryAfter, err := rl.Allow(c.Context(), key, limit, window)
		if err != nil {
			return c.Next()
		}

		if !allowed {
			c.Set("Retry-After", fmt.Sprintf("%d", int(math.Ceil(retryAfter.Seconds()))))
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"code":  http.StatusTooManyRequests,
				"error": "too many requests",
			})
		}

		return c.Next()
	}
}
