package routes

import (
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewCoreSubmissionRoutes(router fiber.Router) {
	submissionRouter := router.Group("/submissions")

	submissionRouter.Post("/", middlewares.ValidateMiddleware[requests.Submission](), func(c *fiber.Ctx) error {
		return nil
	})
}
