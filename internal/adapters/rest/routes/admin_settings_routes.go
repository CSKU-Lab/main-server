package routes

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v3"
)

type adminSettingsRequest struct {
	DefaultCompareScriptID string `json:"default_compare_script_id"`
}

func (r *adminSettingsRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.DefaultCompareScriptID, validation.Required),
	)
}

type adminSettingsResponse struct {
	DefaultCompareScriptID string `json:"default_compare_script_id"`
}

func NewAdminSettingsRoutes(router fiber.Router, settingsService services.SystemSettingsService) {
	settingsRouter := router.Group("/settings", middlewares.RequireAdmin())

	settingsRouter.Get("/", func(c fiber.Ctx) error {
		return c.JSON(&adminSettingsResponse{
			DefaultCompareScriptID: settingsService.GetDefaultCompareScriptID(c.RequestCtx()),
		})
	})

	settingsRouter.Put("/", middlewares.ValidateMiddleware[adminSettingsRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*adminSettingsRequest)
		if err := settingsService.SetDefaultCompareScriptID(c.RequestCtx(), req.DefaultCompareScriptID); err != nil {
			return err
		}
		return c.SendStatus(http.StatusNoContent)
	})
}
