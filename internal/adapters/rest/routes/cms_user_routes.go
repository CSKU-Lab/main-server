package routes

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

func NewCMSUserRoute(router fiber.Router, userService services.UserService) {
	cmsUserRoute := router.Group("/users")

	cmsUserRoute.Get("/:userID", middlewares.RequireAdminOrInstructor(), func(c fiber.Ctx) error {
		userID := c.Params("userID")

		user, err := userService.GetByID(c.RequestCtx(), userID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(&models.CMSUser{
			ID:           user.ID,
			Username:     user.Username,
			DisplayName:  user.DisplayName,
			ProfileImage: user.ProfileImage,
		})
	})
}
