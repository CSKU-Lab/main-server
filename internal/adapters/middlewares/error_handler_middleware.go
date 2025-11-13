package middlewares

import (
	"errors"
	"fmt"
	"log"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/gofiber/fiber/v2"
)

type errorHandlerMiddleware struct {
	appConfig *configs.Config
}

func NewErrorHandlerMiddleware(appConfig *configs.Config) *errorHandlerMiddleware {
	return &errorHandlerMiddleware{
		appConfig: appConfig,
	}
}

func (e *errorHandlerMiddleware) ErrorHandler(c *fiber.Ctx, err error) error {
	log.Println(err)
	var csErr *cserrors.Error
	if errors.As(err, &csErr) {
		var errString string
		if csErr.Code != nil {
			errString = csErr.Code.Error()
		}

		return c.Status(csErr.HttpStatus).JSON(fiber.Map{
			"code":  errString,
			"error": csErr.Message,
		})
	}

	var redirectErr cserrors.RedirectError
	if errors.As(err, &redirectErr) {
		return c.Redirect(fmt.Sprintf("%s/auth/sign-in?error=%s", e.appConfig.FRONTEND_URL, redirectErr.Code()), fiber.StatusTemporaryRedirect)
	}

	if errors.Is(err, fiber.ErrMethodNotAllowed) {
		return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"error": "Method Not Allowed",
		})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		if fiberErr.Code == 404 {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Route not found",
				"message": fiberErr.Message,
			})
		}

	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal Server Error",
	})
}
