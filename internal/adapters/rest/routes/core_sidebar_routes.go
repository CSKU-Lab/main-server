package routes

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/gofiber/fiber/v2"
)

func NewCoreSidebarRoute(router fiber.Router, sidebarService services.SidebarService) {
	coreSidebarRouter := router.Group("/sidebar")

	coreSidebarRouter.Get("/", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		sidebars, err := sidebarService.GetSidebar(c.Context(), user.ID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(sidebars)
	})
}
