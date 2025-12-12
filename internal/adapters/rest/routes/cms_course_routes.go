package routes

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewCMSCourseRoutes(router fiber.Router, courseService services.CourseService, sectionService services.SectionService, semesterService services.SemesterService) {
	courseRouter := router.Group("/courses")

	courseRouter.Get("/:courseID/sections", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "desc")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		filterParams := make(map[string]string)
		for key, value := range c.Queries() {
			if strings.Contains(key, "__") {
				filterParams[key] = value
			}
		}

		sems, err := semesterService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: err.Error()})
		}

		count, err := semesterService.Count(c.Context(), search, nil)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semesters count"})
		}

		type semesterFields struct {
			Name string              `json:"name"`
			Type models.SemesterType `json:"type"`
		}

		type sectionsOfSemester struct {
			Semester semesterFields   `json:"semester"`
			Sections []models.Section `json:"sections"`
		}

		sectionsOfSemesters := make([]sectionsOfSemester, len(sems))
		for i, semester := range sems {
			sections, err := sectionService.GetByCourseIDAndSemesterID(c.Context(), courseID, semester.ID)
			if err != nil {
				return err
			}

			responseSections := []models.Section{}
			if len(sections) > 0 {
				responseSections = sections
			}

			sectionsOfSemesters[i] = sectionsOfSemester{
				Semester: semesterFields{
					Name: semester.Name,
					Type: semester.Type,
				},
				Sections: responseSections,
			}
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
				"total_rows": count,
			},
			"data": sectionsOfSemesters,
		})
	})

	courseRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateCourse](), func(c *fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateCourse)

		course, err := courseService.Create(c.Context(), req)
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

		courses, err := courseService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder, show)
		if err != nil {
			return err
		}

		count, err := courseService.Count(c.Context(), search, show)
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
		course, err := courseService.GetByID(c.Context(), courseID)
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

	courseRouter.Patch("/:courseID", middlewares.ValidateMiddleware[requests.UpdateCourse](), func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")
		course := c.Locals("body").(*requests.UpdateCourse)

		err := courseService.UpdateByID(c.Context(), courseID, course)
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

		err := courseService.DeleteByID(c.Context(), courseID)
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
