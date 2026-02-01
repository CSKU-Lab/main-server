package routes

import (
	"encoding/json"

	"github.com/CSKU-Lab/main-server/domain/models"
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

	type InnerFakePayload struct {
		Code           string                `json:"code"`
		Status         string                `json:"status"`
		AvgWallTime    float32               `json:"avg_wall_time"`
		AvgMemory      int32                 `json:"avg_memory"`
		TestCaseGroups models.TestCaseGroups `json:"test_case_groups"`
	}

	type FakePayload struct {
		Payload InnerFakePayload `json:"payload"`
	}

	submissionRouter.Patch("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		payload := &services.UpdateSubmissionPayload{
			Type:   "code",
			Status: "passed",
			Payload: InnerFakePayload{
				Code:        "updated code",
				Status:      "RUN_PASSED",
				AvgWallTime: 0.1293193,
				AvgMemory:   10,
				TestCaseGroups: models.TestCaseGroups{
					models.TestCaseGroup{
						ID:    "d860de9d-6264-44eb-9aa7-15842142d3a7",
						Score: 100,
						TestCases: []models.TestCase{
							{
								ID:       "9d7ae2df-4726-4893-ba2b-cffd74f3718f",
								Status:   "STATUS_RUN_PASSED",
								Output:   "19\nbusted\n16\n",
								Message:  "",
								WallTime: 0.010999999940395355,
								Memory:   8372,
							},
						},
					},
				},
			},
		}

		rawPayload, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		return service.Update(c.UserContext(), id, payload, rawPayload)
	})
}
