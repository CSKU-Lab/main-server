package middlewares

import (
	"errors"
	"fmt"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/internal/logging"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type ErrorHandlerMiddleware struct {
	appConfig *configs.Config
	logger    *zap.SugaredLogger
}

func NewErrorHandlerMiddleware(appConfig *configs.Config, logger *zap.SugaredLogger) *ErrorHandlerMiddleware {
	return &ErrorHandlerMiddleware{
		appConfig: appConfig,
		logger:    logger,
	}
}

func (e *ErrorHandlerMiddleware) ErrorHandler(c fiber.Ctx, err error) error {
	log := logging.FromContext(c.Context())

	var redirectErr cserrors.RedirectError
	if errors.As(err, &redirectErr) {
		log.Errorw("request error",
			"error", err,
			"redirect_code", redirectErr.Code(),
			"method", c.Method(),
			"path", c.Path(),
		)
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(fmt.Sprintf("%s/auth/sign-in?error=%s", e.appConfig.FRONTEND_URL, redirectErr.Code()))
	}

	log.Errorw("request error",
		"error", err,
		"method", c.Method(),
		"path", c.Path(),
	)

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
