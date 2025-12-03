package routes

import (
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewCMSUserExistancesRoutes(router fiber.Router, userService services.UserService) {
	userRouter := router.Group("/user-existances")

	userRouter.Post("/", middlewares.ValidateMiddleware[requests.GetUserExistances](), func(c *fiber.Ctx) error {
		req := c.Locals("body").(*requests.GetUserExistances)

		res, err := userService.GetUserExistances(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error fetching user existances",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(res)
	})
}
