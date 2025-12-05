package middlewares

import (
	"net/http"
	"slices"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/gofiber/fiber/v2"
)

func RBACMiddleware(roles []models.Role) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		for _, role := range user.Roles {
			if slices.Contains(roles, role) {
				return c.Next()
			}
		}

		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "No Permission"})
	}
}
