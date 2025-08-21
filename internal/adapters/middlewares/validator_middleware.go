package middlewares

import (
	"errors"
	"reflect"

	"github.com/CSKU-Lab/main-server/internal/validator"
	"github.com/gofiber/fiber/v2"
)

func ValidateMiddleware(r any) func(*fiber.Ctx) error {
	if err := isStructPointer(r); err != nil {
		return func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "Invalid using ValidateMiddleware need to pass struct pointer to the middlware",
			})
		}
	}

	appValidator := validator.NewAppValidator()

	return func(c *fiber.Ctx) error {
		c.BodyParser(&r)

		if errs := appValidator.Validate(r); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Bad Request",
				"fields": errs,
			})
		}

		c.Locals("request", r)

		return c.Next()
	}
}

func isStructPointer(i any) error {
	typ := reflect.TypeOf(i)
	if typ.Kind() != reflect.Ptr {
		return errors.New("not a pointer")
	}

	return nil
}
