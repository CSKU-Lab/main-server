package routes

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCMSCourseRoutes(router fiber.Router, courseService services.CourseService, sectionService services.SectionService, semesterService services.SemesterService, defaultLabService services.DefaultLabService, labService services.LabService, permService permission.Service) {
	courseRouter := router.Group("/courses", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	courseRouter.Get("/:courseID/sections", func(c fiber.Ctx) error {
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

		sems, err := semesterService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: err.Error()})
		}

		count, err := semesterService.Count(c.RequestCtx(), search, nil)
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

		sectionsOfSemesters := []sectionsOfSemester{}
		for _, semester := range sems {
			sections, err := sectionService.GetByCourseIDAndSemesterID(c.RequestCtx(), courseID, semester.ID)
			if err != nil {
				return err
			}

			if len(sections) == 0 {
				continue
			}

			sectionsOfSemesters = append(sectionsOfSemesters, sectionsOfSemester{
				Semester: semesterFields{
					Name: semester.Name,
					Type: semester.Type,
				},
				Sections: sections,
			})
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

	courseRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateCourse](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateCourse)

		course, err := courseService.Create(c.RequestCtx(), req)
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

	courseRouter.Get("/", func(c fiber.Ctx) error {
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

		courses, err := courseService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, show)
		if err != nil {
			return err
		}

		count, err := courseService.Count(c.RequestCtx(), search, show)
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

	courseRouter.Get("/:courseID", func(c fiber.Ctx) error {
		courseID := c.Params("courseID")
		course, err := courseService.GetByID(c.RequestCtx(), courseID)
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

	courseRouter.Patch("/:courseID", middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.UpdateCourse](), func(c fiber.Ctx) error {
		courseID := c.Params("courseID")
		course := c.Locals("body").(*requests.UpdateCourse)

		err := courseService.UpdateByID(c.RequestCtx(), courseID, course)
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

	courseRouter.Delete("/:courseID", middlewares.Permission(permService).ForCourse("courseID").CanDelete(), func(c fiber.Ctx) error {
		courseID := c.Params("courseID")

		err := courseService.DeleteByID(c.RequestCtx(), courseID)
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

	courseRouter.Post("/:courseID/default-labs", middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.SetDefaultLab](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")
		req := c.Locals("body").(*requests.SetDefaultLab)
		err := defaultLabService.Create(c.RequestCtx(), req, user.ID, courseID)
		if err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusCreated)
	})

	courseRouter.Get("/:courseID/default-labs", func(c fiber.Ctx) error {
		courseID := c.Params("courseID")

		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "position")
		sortOrder := c.Query("sort_order", "asc")

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

		filterParams["course_id__is"] = courseID

		defaultLabs, err := defaultLabService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := defaultLabService.Count(c.RequestCtx(), filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": defaultLabs,
		})
	})

	courseRouter.Post("/:courseID/default-labs/delete", middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.DeleteDefaultLab](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")
		req := c.Locals("body").(*requests.DeleteDefaultLab)
		return defaultLabService.Delete(c.RequestCtx(), req, user.ID, courseID)
	})

	courseRouter.Patch("/:courseID/default-labs", middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.UpdateDefaultLab](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")
		req := c.Locals("body").(*requests.UpdateDefaultLab)
		return defaultLabService.Update(c.RequestCtx(), req, user.ID, courseID)
	})

	courseRouter.Get("/:courseID/labs", func(c fiber.Ctx) error {
		courseID := c.Params("courseID")

		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "display_name")
		sortOrder := c.Query("sort_order", "asc")

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

		filterParams["course_id__is"] = courseID

		labs, err := labService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labService.Count(c.RequestCtx(), search, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": labs,
		})
	})
}

// fiber:context-methods migrated
