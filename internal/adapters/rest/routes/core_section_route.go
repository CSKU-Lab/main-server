package routes

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/gofiber/fiber/v3"
)

func NewCoreSectionRoute(router fiber.Router, sectionService services.SectionService, labSectionService services.LabSectionService, labService services.LabService, sectionStudentService services.SectionStudentService, labMaterialService services.LabMaterialService, courseService services.CourseService) {
	coreSectionRouter := router.Group("/sections")

	coreSectionRouter.Get("/", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		// search := c.Query("search", "")
		sortBy := c.Query("sort_by", "created_at")
		sortOrder := c.Query("sort_order", "desc")
		user := c.Locals("user").(*models.User)

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

		filterParams["student_id__is"] = user.ID

		sections, err := sectionService.GetSectionsPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: err.Error()})
		}

		count, err := sectionService.Count(c.RequestCtx(), filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting sections count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
				"total_rows": count,
			},
			"data": sections,
		})
	})

	coreSectionRouter.Get("/:sectionID", func(c fiber.Ctx) error {
		sectionID := c.Params("sectionID")
		user := c.Locals("user").(*models.User)

		secStudent, err := sectionStudentService.GetBySectionAndStudentID(c.RequestCtx(), sectionID, user.ID)
		if err != nil {
			return err
		}

		section, err := sectionService.GetByID(c.RequestCtx(), secStudent.SectionID)
		if err != nil {
			return err
		}
		course, err := courseService.GetByID(c.RequestCtx(), section.CourseID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"section": section,
			"course":  course,
		})
	})

	coreSectionRouter.Get("/:sectionID/labs/:labID", func(c fiber.Ctx) error {
		sectionID := c.Params("sectionID")
		labID := c.Params("labID")
		user := c.Locals("user").(*models.User)

		secStudent, err := sectionStudentService.GetBySectionAndStudentID(c.RequestCtx(), sectionID, user.ID)
		if err != nil {
			return err
		}

		labSec, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, secStudent.SectionID)
		if err != nil {
			return err
		}

		return c.JSON(labSec)
	})

	coreSectionRouter.Get("/:sectionID/labs", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		// search := c.Query("search", "")
		sortBy := c.Query("sort_by", "position")
		sortOrder := c.Query("sort_order", "asc")

		user := c.Locals("user").(*models.User)
		sectionID := c.Params("sectionID")

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

		secStudent, err := sectionStudentService.GetBySectionAndStudentID(c.RequestCtx(), sectionID, user.ID)
		if err != nil {
			return err
		}

		filterParams["section_id__is"] = secStudent.SectionID

		labSections, err := labSectionService.GetPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labSectionService.Count(c.RequestCtx(), filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
		}

		type labSectionResponse struct {
			ID        string `json:"id"`
			LabID     string `json:"lab_id"`
			SectionID string `json:"section_id"`
			Position  int    `json:"position"`
			LabName   string `json:"lab_name"`
		}

		responseSections := make([]labSectionResponse, len(labSections))
		for i, section := range labSections {
			lab, err := labService.GetByID(c.RequestCtx(), section.LabID)
			if err != nil {
				return err
			}

			responseSections[i] = labSectionResponse{
				ID:        section.ID,
				LabID:     section.LabID,
				SectionID: section.SectionID,
				Position:  section.Position,
				LabName:   lab.DisplayName,
			}
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": responseSections,
		})
	})

	coreSectionRouter.Delete("/:sectionID/unenroll", func(c fiber.Ctx) error {
		sectionID := c.Params("sectionID")
		user := c.Locals("user").(*models.User)

		secStudent, err := sectionStudentService.GetBySectionAndStudentID(c.RequestCtx(), sectionID, user.ID)
		if err != nil {
			return err
		}

		ctx := c.Context()
		err = sectionService.RemoveStudents(ctx, secStudent.SectionID, []string{secStudent.StudentID})
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Students removed successfully",
		})
	})
}

// fiber:context-methods migrated
