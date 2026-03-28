package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission creates a middleware that checks if the user satisfies the given permission condition.
// The condition is evaluated against the user ID extracted from the JWT token stored in context locals.
//
// Example usage:
//
//	router.Get("/", RequirePermission(
//	    permission.Or(
//	        permission.IsAdmin,
//	        permission.IsSectionInstructor("id"),
//	    ),
//	), handler)
func RequirePermission(condition permission.Condition) func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		err := permission.User(user.ID).
			Conditions(condition).
			Check(c.RequestCtx())

		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "Forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}
