package routes

import (
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewCoreSubmissionRoutes(router fiber.Router, service services.SubmissionService) {
	submissionRouter := router.Group("/submissions")

	submissionRouter.Post("/", middlewares.ValidateMiddleware[requests.Submission](), func(c *fiber.Ctx) error {
		payload := c.Locals("body").(*requests.Submission)

		id, err := service.Create(c.UserContext(), payload, c.Body())
		if err != nil {
			return err
		}

		return c.JSON(&fiber.Map{
			"id": id,
		})
	})

	submissionRouter.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		submission, err := service.Get(c.Context(), id)
		if err != nil {
			return err
		}

		return c.JSON(submission)
	})
}
