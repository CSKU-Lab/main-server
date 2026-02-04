package routes

import (
	"bufio"
	"context"
	"fmt"
	"log"

	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCoreSubmissionRoutes(router fiber.Router, service services.SubmissionService, ps pubsub.PubSub) {
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

	submissionRouter.Get("/listen/:id", func(c fiber.Ctx) error {
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

			channel := fmt.Sprintf("submissions:update:%s", id)

			err = ps.Subscribe(ctx, channel, func(payload []byte) error {
				status := string(payload)

				fmt.Fprintf(w, "data: ID: %s, Status: %s\n\n", id, status)
				err := w.Flush()
				if err != nil {
					cancel()
					return err
				}

				if status == "failed" || status == "passed" {
					return pubsub.Exit
				}
				return nil
			})
			if err != nil {
				log.Println(err)
			}
		})
		return nil
	})
}

// fiber:context-methods migrated
