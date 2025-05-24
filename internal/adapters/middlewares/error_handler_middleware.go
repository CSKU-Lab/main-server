package middlewares

import (
	"errors"
	"fmt"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
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
	var csErr *cserrors.Error
	if errors.As(err, &csErr) {
		return c.Status(int(csErr.Code)).JSON(fiber.Map{
			"code":  csErr.Code,
			"error": csErr.Message,
		})
	}

	var redirectErr cserrors.RedirectError
	if errors.As(err, &redirectErr) {
		return c.Redirect(fmt.Sprintf("%s/auth/sign-in?error=%s", e.appConfig.FRONTEND_URL, redirectErr.Code()), fiber.StatusTemporaryRedirect)
	}

	if errors.Is(err, fiber.ErrMethodNotAllowed) {
		return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"code":  "METHOD_NOT_ALLOWED",
			"error": "Method Not Allowed",
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code":  "INTERNAL_SERVER_ERROR",
		"error": "Internal Server Error",
	})
}
