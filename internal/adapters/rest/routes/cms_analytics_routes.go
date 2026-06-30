package routes

import (
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

const (
	defaultCMSAnalyticsDays = 30
	maxCMSAnalyticsDays     = 365
)

// NewCMSAnalyticsRoutes exposes the same global analytics overview as the admin
// dashboard, but to the CMS home page so instructors (and admins) get the
// platform-wide daily stats on landing. Data is identical to /admin/analytics.
func NewCMSAnalyticsRoutes(router fiber.Router, analyticsService services.AnalyticsService) {
	analyticsRouter := router.Group("/analytics", middlewares.RequireAdminOrInstructor())

	analyticsRouter.Get("/overview", func(c fiber.Ctx) error {
		days := defaultCMSAnalyticsDays
		if raw := c.Query("days"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				days = parsed
			}
		}
		if days > maxCMSAnalyticsDays {
			days = maxCMSAnalyticsDays
		}

		overview, err := analyticsService.GetOverview(c.RequestCtx(), days)
		if err != nil {
			return err
		}
		return c.JSON(overview)
	})
}
