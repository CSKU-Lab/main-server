package middlewares

import (
	"net/http"
	"slices"

	"github.com/CSKU-Lab/main-server/constants"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/gofiber/fiber/v2"
)

func AdminMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	if slices.Contains(user.Roles, constants.ADMIN) {
		return c.Next()
	}

	return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
}
