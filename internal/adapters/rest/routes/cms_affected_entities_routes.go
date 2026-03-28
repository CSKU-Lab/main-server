package routes

import (
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCMSAffectedEntitiesRoutes(router fiber.Router, affectedEntitiesService services.AffectedEntitiesService) {
	affectedEntitiesRouter := router.Group("/affected-entities")

	// POST /api/v1/cms/affected-entities - Get affected entities (Admin only)
	affectedEntitiesRouter.Post("/",
		middlewares.RequirePermission(permission.IsAdmin),
		middlewares.ValidateMiddleware[requests.GetAffectedEntities](),
		func(c fiber.Ctx) error {
			req := c.Locals("body").(*requests.GetAffectedEntities)

			res, err := affectedEntitiesService.GetAffectedEntities(c.RequestCtx(), req)
			if err != nil {
				return err
			}

			return c.Status(fiber.StatusOK).JSON(res)
		})
}

// fiber:context-methods migrated
