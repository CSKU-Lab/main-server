package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission creates a middleware that checks if the user satisfies the given permission condition.
// It retrieves the user from the context (set by ProtectedRouteMiddleware) and validates the condition.
//
// Example usage:
//
//	router.Get("/admin/users", RequirePermission(permission.IsAdmin), handler.GetUsers)
//	router.Get("/me", RequirePermission(permission.IsAuthenticated), handler.GetMe)
//
// The middleware chain should include ProtectedRouteMiddleware before RequirePermission:
//
//	router.Use(ProtectedRouteMiddleware(secret))
//	router.Get("/admin", RequirePermission(permission.IsAdmin), handler.AdminHandler)
func RequirePermission(condition permission.Condition) func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		if !condition.IsSatisfied(user.ID) {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "Forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}
