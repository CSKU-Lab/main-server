package routes

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"

	"github.com/gofiber/fiber/v3"
)

func NewCMSSubmissionRoutes(router fiber.Router, submissionService services.SubmissionService, permService permission.Service) {
	submissionRouter := router.Group("/submissions")

	// Get submission by ID - instructors can view any submission in their sections
	submissionRouter.Get("/:id",
		middlewares.Permission(permService).ForSubmission("id").CanView(),
		func(c fiber.Ctx) error {
			id := c.Params("id")

			submission, err := submissionService.Get(c.RequestCtx(), id)
			if err != nil {
				return err
			}

			return c.JSON(submission)
		},
	)

	// Grade submission - instructors can grade submissions
	submissionRouter.Post("/:id/grade",
		middlewares.Permission(permService).ForSubmission("id").CanGrade(),
		middlewares.ValidateMiddleware[requests.UpdateSubmissionManualScore](),
		func(c fiber.Ctx) error {
			id := c.Params("id")
			req := c.Locals("body").(*requests.UpdateSubmissionManualScore)

			if err := submissionService.UpdateManualScore(c.RequestCtx(), id, req.ManualScore); err != nil {
				return err
			}

			return c.SendStatus(http.StatusNoContent)
		},
	)

	// Delete submission - instructors can delete any submission in their sections
	submissionRouter.Delete("/:id",
		middlewares.Permission(permService).ForSubmission("id").CanGrade(),
		func(c fiber.Ctx) error {
			id := c.Params("id")
			if err := submissionService.DeleteSubmission(c.RequestCtx(), id); err != nil {
				return err
			}
			return c.SendStatus(http.StatusNoContent)
		},
	)

	// Update manual score (grade) - instructors can grade submissions
	submissionRouter.Patch("/:id/manual-score",
		middlewares.Permission(permService).ForSubmission("id").CanGrade(),
		middlewares.ValidateMiddleware[requests.UpdateSubmissionManualScore](),
		func(c fiber.Ctx) error {
			id := c.Params("id")
			req := c.Locals("body").(*requests.UpdateSubmissionManualScore)

			if err := submissionService.UpdateManualScore(c.RequestCtx(), id, req.ManualScore); err != nil {
				return err
			}

			return c.SendStatus(http.StatusNoContent)
		},
	)
}
