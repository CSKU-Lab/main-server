package routes

import (
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

const (
	defaultAnalyticsDays = 30
	maxAnalyticsDays     = 365
)

func NewAdminAnalyticsRoutes(router fiber.Router, analyticsService services.AnalyticsService) {
	analyticsRouter := router.Group("/analytics", middlewares.RequireAdmin())

	analyticsRouter.Get("/overview", func(c fiber.Ctx) error {
		days := defaultAnalyticsDays
		if raw := c.Query("days"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				days = parsed
			}
		}
		if days > maxAnalyticsDays {
			days = maxAnalyticsDays
		}

		overview, err := analyticsService.GetOverview(c.RequestCtx(), days)
		if err != nil {
			return err
		}
		return c.JSON(overview)
	})
}
