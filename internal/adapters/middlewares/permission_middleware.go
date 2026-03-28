package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission creates a middleware that checks if the user satisfies the given permission condition.
// It extracts the user from the context (set by authentication middleware) and validates the permission.
//
// Example usage:
//
//	router.Get("/", middlewares.RequirePermission(permission.IsAdmin), handler)
//	router.Get("/", middlewares.RequirePermission(permission.Or(permission.IsAdmin, permission.IsInstructor)), handler)
func RequirePermission(condition permission.Condition) func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		err := permission.User(user.ID).
			Conditions(condition).
			Check()

		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "Forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}
