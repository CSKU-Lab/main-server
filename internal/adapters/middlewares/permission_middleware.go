package middlewares

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission creates a middleware that checks if the user satisfies the given permission condition.
// It extracts the user from the context and validates the condition.
// If the condition is not satisfied, it returns a 403 Forbidden error.
//
// Example usage:
//
//	router.Get("/:id",
//		middlewares.RequirePermission(
//			permission.Or(
//				permission.IsAdmin,
//				permission.IsSectionInstructor("id"),
//				permission.IsSectionStudent("id"),
//			),
//		),
//		handler,
//	)
func RequirePermission(condition permission.Condition) func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		err := permission.User(user.ID).
			Conditions(condition).
			CheckWithContextAndHTTPError(c.Context())

		if err != nil {
			return err
		}

		return c.Next()
	}
}
