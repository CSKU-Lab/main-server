package routes

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewAdminSemesterRoutes(router fiber.Router, service services.SemesterService, sectionService services.SectionService, courseService services.CourseService) {
	semesterRouter := router.Group("/semesters", middlewares.RequireAdmin())

	semesterRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateSemester](), func(c fiber.Ctx) error {
		sem := c.Locals("body").(*requests.CreateSemester)

		err := service.Create(c.RequestCtx(), sem)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error creating semester"})
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	semesterRouter.Get("/", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "desc")

		filterParams := make(map[string]string)
		for key, value := range c.Queries() {
			if strings.Contains(key, "__") {
				filterParams[key] = value
			}
		}

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		sems, err := service.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := service.Count(c.RequestCtx(), search, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semesters count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
				"total_rows": count,
			},
			"data": sems,
		})
	})

	semesterRouter.Get("/:semID", func(c fiber.Ctx) error {
		semID := c.Params("semID")
		sem, err := service.GetByID(c.RequestCtx(), semID)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semester"})
		}

		return c.JSON(sem)
	})

	semesterRouter.Patch("/:semID", middlewares.ValidateMiddleware[requests.UpdateSemester](), func(c fiber.Ctx) error {
		ID := c.Params("semID")

		sem := c.Locals("body").(*requests.UpdateSemester)

		err := c.Bind().Body(sem)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Error parsing request"})
		}

		err = service.UpdateByID(c.RequestCtx(), ID, sem)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error updating semester"})
		}

		return c.SendStatus(fiber.StatusAccepted)
	})

	semesterRouter.Delete("/:semID", func(c fiber.Ctx) error {
		err := service.DeleteByID(c.RequestCtx(), c.Params("semID"))
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error deleting semester"})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	semesterRouter.Get("/:semID/affected-sections", func(c fiber.Ctx) error {
		semID := c.Params("semID")

		courseWithSections, err := service.GetAffectedSections(c.RequestCtx(), semID)
		if err != nil {
			return err
		}

		return c.JSON(courseWithSections)
	})
}

// fiber:context-methods migrated
