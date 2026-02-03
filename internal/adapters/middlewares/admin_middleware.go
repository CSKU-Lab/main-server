package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/gofiber/fiber/v3"
)

func AdminMiddleware(c fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	for _, role := range user.Roles {
		if role == models.ADMIN {
			return c.Next()
		}
	}

	return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
}
