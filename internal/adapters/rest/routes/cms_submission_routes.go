package routes

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"

	"github.com/gofiber/fiber/v3"
)

func NewCMSSubmissionRoutes(router fiber.Router, submissionService services.SubmissionService) {
	submissionRouter := router.Group("/submissions", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	submissionRouter.Patch("/:id/manual-score", middlewares.ValidateMiddleware[requests.UpdateSubmissionManualScore](), func(c fiber.Ctx) error {
		id := c.Params("id")
		req := c.Locals("body").(*requests.UpdateSubmissionManualScore)

		if err := submissionService.UpdateManualScore(c.RequestCtx(), id, req.ManualScore); err != nil {
			return err
		}

		return c.SendStatus(http.StatusNoContent)
	})
}
