package routes

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v2"
)

func NewCMSCourseRoutes(router fiber.Router, sectionService services.SectionService, semesterService services.SemesterService) {
	courseRouter := router.Group("/course")

	courseRouter.Get("/:courseID/sections", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
		models.STUDENT,
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
			sections, err := sectionService.GetByCourseID(c.Context(), courseID)
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
}
