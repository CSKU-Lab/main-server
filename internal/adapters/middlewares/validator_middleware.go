package middlewares

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v3"
)

func ValidateMiddleware[T any, PT interface {
	*T
	validation.Validatable
}]() func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		body := new(T)
		err := c.Bind().Body(body)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid request body",
			})
		}

		var validator PT = body
		if err := validator.Validate(); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Bad Request",
				"fields": err,
			})
		}

		c.Locals("body", body)

		return c.Next()
	}
}
