package routes

import (
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCMSUserExistancesRoutes(router fiber.Router, userService services.UserService) {
	userRouter := router.Group("/user-existances")

	// POST /api/v1/cms/user-existances - Check user existances (Admin or Instructor)
	userRouter.Post("/",
		middlewares.RequirePermission(permission.Or(permission.IsAdmin, permission.IsInstructor)),
		middlewares.ValidateMiddleware[requests.GetInvalidUsers](),
		func(c fiber.Ctx) error {
			req := c.Locals("body").(*requests.GetInvalidUsers)

			res, err := userService.GetInvalidUsers(c.RequestCtx(), req)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Error fetching user existances",
					"error":   err.Error(),
				})
			}

			if res != nil {
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"code":  "INVALID_USERS",
					"error": "Some users are invalid",
					"users": res,
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"code":    "OK",
				"message": "All users are valid",
			})
		})
}

// fiber:context-methods migrated
