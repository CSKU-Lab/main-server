package routes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCoreSubmissionRoutes(router fiber.Router, service services.SubmissionService) {
	submissionRouter := router.Group("/submissions")

	submissionRouter.Post("/", middlewares.ValidateMiddleware[requests.Submission](), func(c fiber.Ctx) error {
		payload := c.Locals("body").(*requests.Submission)

		id, err := service.Create(c.Context(), payload, c.Body())
		if err != nil {
			return err
		}

		return c.JSON(&fiber.Map{
			"id": id,
		})
	})

	submissionRouter.Get("/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		submission, err := service.Get(c.RequestCtx(), id)
		if err != nil {
			return err
		}

		return c.JSON(submission)
	})
	type submission struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	submissionRouter.Get("/:id/listen", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		id := c.Params("id")

		ctx, cancel := context.WithCancel(context.Background())

		c.Status(fiber.StatusOK).RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
			fmt.Fprint(w, "event: connected\ndata: connected\n\n")
			err := w.Flush()
			if err != nil {
				cancel()
				log.Println(err)
			}

			submissionChan, err := service.Listen(ctx, id)
			if err != nil {
				cancel()
				log.Println(err)
			}

			for sub := range submissionChan {
				data, err := json.Marshal(sub)
				if err != nil {
					cancel()
					log.Println(err)
				}

				fmt.Fprintf(w, "data: %s\n\n", data)
				err = w.Flush()
				if err != nil {
					cancel()
					log.Println(err)
				}
			}

		})
		return nil
	})
}

// fiber:context-methods migrated
