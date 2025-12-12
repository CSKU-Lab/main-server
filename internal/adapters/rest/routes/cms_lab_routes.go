package routes

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewCMSLabRoutes(router fiber.Router, labService services.LabService) {
	labRouter := router.Group("/labs")

	labRouter.Get("/:labID", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c *fiber.Ctx) error {
		labID := c.Params("labID")
		lab, err := labService.GetByID(c.Context(), labID)
		if err != nil {
			return err
		}
		return c.JSON(lab)
	})

	labRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateLab](), middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c *fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateLab)

		user := c.Locals("user").(*models.User)

		labID, err := labService.Create(c.Context(), req, user.ID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": labID,
		})
	})
}
