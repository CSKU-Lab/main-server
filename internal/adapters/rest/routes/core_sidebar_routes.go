package routes

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

func NewCoreSidebarRoute(router fiber.Router, sidebarService services.SidebarService) {
	coreSidebarRouter := router.Group("/sidebar")

	coreSidebarRouter.Get("/", middlewares.RequireAuthenticated(), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		sidebars, err := sidebarService.GetSidebar(c.RequestCtx(), user.ID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(sidebars)
	})
}

// fiber:context-methods migrated
