package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
)

// RequirePermission creates a middleware that checks if the user satisfies the given condition.
// If the condition is not satisfied, it returns a 403 Forbidden error.
//
// Example usage:
//
//	permService := permission.NewPermissionService(...)
//	router.Get("/courses/:id",
//	    middlewares.ProtectedRouteMiddleware(config.JWTSecret),
//	    middlewares.RequirePermission(permService.IsCourseCreator("id")),
//	    handler.GetCourse)
//
// Complex conditions:
//
//	router.Get("/admin",
//	    middlewares.ProtectedRouteMiddleware(config.JWTSecret),
//	    middlewares.RequirePermission(
//	        permService.Or(
//	            permService.IsAdmin(),
//	            permService.IsCourseCreator("id"),
//	        ),
//	    ),
//	    handler.AdminHandler)
func RequirePermission(condition permission.Condition) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user from context (set by ProtectedRouteMiddleware)
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusUnauthorized,
				Message:    "Authentication required",
			})
		}

		// Build params map from route parameters
		params := extractParams(c)

		// Evaluate the condition
		if !condition.Evaluate(c.Context(), user, params) {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "Access denied",
			})
		}

		return c.Next()
	}
}

// extractParams extracts route parameters from the Fiber context.
// It uses the route's Params definition to get all parameter names,
// then retrieves each value using c.Params().
func extractParams(c fiber.Ctx) map[string]string {
	params := make(map[string]string)

	// Get the matched route
	route := c.Route()
	if route == nil {
		return params
	}

	// Iterate over the route's parameter names and get their values
	for _, paramName := range route.Params {
		if paramName != "" {
			value := c.Params(paramName)
			if value != "" {
				params[paramName] = value
			}
		}
	}

	return params
}

// RequireAnyPermission creates a middleware that checks if the user satisfies ANY of the given conditions.
// This is a convenience wrapper around RequirePermission with Or logic.
func RequireAnyPermission(permService permission.Service, conditions ...permission.Condition) fiber.Handler {
	return RequirePermission(permService.Or(conditions...))
}

// RequireAllPermissions creates a middleware that checks if the user satisfies ALL of the given conditions.
// This is a convenience wrapper around RequirePermission with And logic.
func RequireAllPermissions(permService permission.Service, conditions ...permission.Condition) fiber.Handler {
	return RequirePermission(permService.And(conditions...))
}
