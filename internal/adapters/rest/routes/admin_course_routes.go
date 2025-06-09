package routes

import (
	"errors"
	"math"
	"strconv"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/middlewares"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewAdminCourseRoutes(router fiber.Router, service services.CourseService) {
	courseRouter := router.Group("/courses")

	courseRouter.Post("/", middlewares.ValidateMiddleware(&models.Course{}), func(c *fiber.Ctx) error {
		var courseRequest requests.Course
		err := c.BodyParser(&courseRequest)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid request")
		}

		user := c.Locals("user").(*models.User)

		createdCourse, err := service.Create(c.Context(), &courseRequest, user.ID)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error creating course")
		}

		return c.Status(fiber.StatusCreated).JSON(createdCourse)

	})

	courseRouter.Get("/", func(c *fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("pageSize", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "created_at")
		sortOrder := c.Query("sort_order", "desc")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid page")
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid page size")
		}

		courses, err := service.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting courses")
		}

		count, err := service.Count(c.Context(), search)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting courses count")
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
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting course")
		}

		return c.JSON(course)
	})

	courseRouter.Patch("/:courseID", func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")
		var course requests.Course

		err := c.BodyParser(&course)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Error parsing request")
		}

		updateCourse, err := service.UpdateByID(c.Context(), courseID, &course)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error updating course")
		}

		return c.JSON(updateCourse)
	})

	courseRouter.Delete("/:courseID", func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")

		err := service.DeleteByID(c.Context(), courseID)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error deleting course")

		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
