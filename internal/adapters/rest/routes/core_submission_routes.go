package routes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCoreSubmissionRoutes(router fiber.Router, service services.SubmissionService, labSectionService services.LabSectionService, labMaterialService services.LabMaterialService, permService permission.Service) {
	submissionRouter := router.Group("/submissions")

	submissionRouter.Post("/", middlewares.ValidateMiddleware[requests.Submission](), func(c fiber.Ctx) error {
		payload := c.Locals("body").(*requests.Submission)
		if payload.SectionID != nil {
			c.Locals("section_id", *payload.SectionID)
		}
		return c.Next()
	}, middlewares.Permission(permService).ForSection("section_id").FromLocals().CanCreate(), func(c fiber.Ctx) error {
		payload := c.Locals("body").(*requests.Submission)

		id, err := service.Create(c.Context(), payload, c.Body())
		if err != nil {
			return err
		}

		return c.JSON(&fiber.Map{
			"id": id,
		})
	})

	submissionRouter.Get("/:id", middlewares.Permission(permService).ForSubmission("id").CanView(), func(c fiber.Ctx) error {
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

	submissionRouter.Get("/:id/listen", middlewares.Permission(permService).ForSubmission("id").CanView(), func(c fiber.Ctx) error {
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
				return
			}

			submissionChan, err := service.Listen(ctx, id)
			if err != nil {
				cancel()
				log.Println("Error listening submission :", err)
				return
			}

			for sub := range submissionChan {
				data, err := json.Marshal(sub)
				if err != nil {
					log.Println(err)
					return
				}

				fmt.Fprintf(w, "data: %s\n\n", data)
				err = w.Flush()
				if err != nil {
					cancel()
					log.Println(err)
					return
				}
			}
		})
		return nil
	})

	submissionRouter.Get("/", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		materialID := c.Query("material_id", "")
		sectionID := c.Query("section_id", "")
		labID := c.Query("lab_id", "")

		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "20")
		sortOrder := c.Query("sort_order", "desc")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}

		submissions, count, err := service.GetUserSubmissions(c.RequestCtx(), user.ID, materialID, labID, sectionID, page, pageSize, sortOrder)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": int(math.Ceil(float64(count) / float64(pageSize))),
				"total_rows": count,
			},
			"data": submissions,
		})
	})
}

// fiber:context-methods migrated
