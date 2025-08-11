package middlewares

import (
	"net/http"
	"slices"

	"github.com/SornchaiTheDev/cs-lab-backend/constants"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/gofiber/fiber/v2"
)

func AdminMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	if slices.Contains(user.Roles, constants.ADMIN) {
		return c.Next()
	}

	return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
}
