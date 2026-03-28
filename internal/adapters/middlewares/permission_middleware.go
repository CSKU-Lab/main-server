package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission creates a middleware that checks if the user has the required permissions.
// It uses the permission package's fluent API to evaluate conditions.
//
// Example usage:
//
//	router.Get("/:id", middlewares.RequirePermission(
//		perm.Or(
//			perm.IsAdmin,
//			perm.IsSectionInstructor("id"),
//		),
//	), handler)
//
// The middleware will:
// 1. Extract the user from the context (set by authentication middleware)
// 2. Evaluate all permission conditions
// 3. Return 403 Forbidden if any condition fails
// 4. Call the next handler if all conditions pass
func RequirePermission(conditions ...permission.Condition) func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		// Build permission check with all conditions (AND logic)
		builder := permission.User(user.ID)
		if len(conditions) > 0 {
			builder.Conditions(conditions...)
		}

		// Check permissions with request context
		if err := builder.Check(c.Context()); err != nil {
			if err == permission.ErrForbidden {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusForbidden,
					Message:    "Forbidden: insufficient permissions",
				})
			}
			return err
		}

		return c.Next()
	}
}
