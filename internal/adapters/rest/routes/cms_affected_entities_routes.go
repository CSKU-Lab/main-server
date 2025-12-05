package routes

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewCMSAffectedEntitiesRoutes(router fiber.Router, affectedEntitiesService services.AffectedEntitiesService) {
	affectedEntitiesRouter := router.Group("/affected-entities", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	affectedEntitiesRouter.Post("/", middlewares.ValidateMiddleware[requests.GetAffectedEntities](), func(c *fiber.Ctx) error {
		req := c.Locals("body").(*requests.GetAffectedEntities)

		res, err := affectedEntitiesService.GetAffectedEntities(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error fetching affected entities",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(res)
	})
}
