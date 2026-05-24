package routes

import (
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v3"
)

func NewLspTokenRoute(router fiber.Router, jwtSecret string) {
	router.Get("/lsp/token", middlewares.RequireAuthenticated(), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.ID,
			"iss": "cs-lab-lsp",
			"exp": time.Now().Add(5 * time.Minute).Unix(),
		})

		signed, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return fiber.ErrInternalServerError
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": signed})
	})
}
