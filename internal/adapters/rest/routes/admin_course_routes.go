package routes

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewAdminCourseRoutes(router fiber.Router, service services.CourseService) {
	courseRouter := router.Group("/courses")

	courseRouter.Post("/" /*middlewares.ValidateMiddleware(&requests.Course{}),*/, func(c *fiber.Ctx) error {
		req := c.Locals("request")

		course, err := service.Create(c.Context(), req.(*requests.Course))
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error creating course",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(course)
	})

	courseRouter.Get("/", func(c *fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "created_at")
		sortOrder := c.Query("sort_order", "desc")
		show := c.Query("show", "active")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid page",
			})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid page size",
			})
		}

		courses, err := service.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder, show)
		if err != nil {
			return err
		}

		count, err := service.Count(c.Context(), search, show)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": courses,
		})
	})

	courseRouter.Get("/:courseID", func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")
		course, err := service.GetByID(c.Context(), courseID)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error getting course",
			})
		}

		return c.JSON(course)
	})

	courseRouter.Patch("/:courseID" /*middlewares.ValidateMiddleware(&requests.Course{}),*/, func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")
		course := c.Locals("request").(*requests.Course)

		err := service.UpdateByID(c.Context(), courseID, course)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error updating course",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	courseRouter.Delete("/:courseID", func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")

		err := service.DeleteByID(c.Context(), courseID)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error deleting course",
			})

		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
