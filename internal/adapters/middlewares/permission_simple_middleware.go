package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/gofiber/fiber/v3"
)

// RequireAdmin returns a middleware that requires the user to have admin role.
// It checks the user roles stored in c.Locals("user") and returns 403 Forbidden
// if the user is not an admin.
func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		for _, role := range user.Roles {
			if role == models.ADMIN {
				return c.Next()
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    "Permission denied: admin access required",
		})
	}
}

// RequireAdminOrInstructor returns a middleware that requires the user to have
// either admin or instructor role. It returns 403 Forbidden if the user has
// neither role.
func RequireAdminOrInstructor() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		for _, role := range user.Roles {
			if role == models.ADMIN || role == models.INSTRUCTOR {
				return c.Next()
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    "Permission denied: admin or instructor access required",
		})
	}
}

// RequireAuthenticated returns a middleware that requires the user to be authenticated.
// This middleware assumes that the ProtectedRouteMiddleware has already run and
// set the user in c.Locals("user"). It returns 401 Unauthorized if no user is found.
func RequireAuthenticated() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusUnauthorized,
				Message:    "Authentication required",
			})
		}

		// Verify it's a valid user model
		_, ok := user.(*models.User)
		if !ok {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusUnauthorized,
				Message:    "Invalid user data",
			})
		}

		return c.Next()
	}
}
