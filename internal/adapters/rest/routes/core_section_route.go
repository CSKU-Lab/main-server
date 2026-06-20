package routes

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

func NewCoreSectionRoute(router fiber.Router, sectionService services.SectionService, labSectionService services.LabSectionService, labService services.LabService, sectionStudentService services.SectionStudentService, labMaterialService services.LabMaterialService, courseService services.CourseService, submissionService services.SubmissionService, permService permission.Service) {
	coreSectionRouter := router.Group("/sections")

	coreSectionRouter.Get("/:sectionID", middlewares.Permission(permService).ForSection("sectionID").CanView(), func(c fiber.Ctx) error {
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

	coreSectionRouter.Get("/:sectionID/labs/:labID", middlewares.Permission(permService).ForSection("sectionID").CanView(), func(c fiber.Ctx) error {
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

		lab, err := labService.GetByID(c.RequestCtx(), labSec.LabID)
		if err != nil {
			return err
		}

		totalMaterials, err := labMaterialService.Count(c.RequestCtx(), map[string]string{"lab_id__is": labSec.LabID})
		if err != nil {
			return err
		}

		allMaterials, err := labMaterialService.GetByLabID(c.RequestCtx(), labSec.LabID)
		if err != nil {
			return err
		}

		completedMaterials := 0
		hasAnySubmission := false
		for _, mat := range allMaterials {
			status := submissionService.GetMaterialStudentStatus(c.RequestCtx(), user.ID, mat.MaterialID, labSec.LabID, labSec.SectionID)
			if status != "not_started" {
				hasAnySubmission = true
				if status == "passed" {
					completedMaterials++
				}
			}
		}

		studentStatus := "not_started"
		if totalMaterials > 0 && completedMaterials == totalMaterials {
			studentStatus = "passed"
		} else if completedMaterials > 0 {
			studentStatus = "in_progress"
		} else if hasAnySubmission {
			studentStatus = "not_passed"
		}

		return c.JSON(models.CoreLabResponse{
			Name:          lab.DisplayName,
			Status:        labSec.EffectiveStatus(),
			ReadonlyAt:    labSec.ReadonlyAt,
			StudentStatus: studentStatus,
		})
	})

	coreSectionRouter.Get("/:sectionID/labs", middlewares.Permission(permService).ForSection("sectionID").CanView(), func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
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
		filterParams["status__is_not"] = "hidden"
		if search != "" {
			filterParams["l.display_name__contains"] = search
		}

		labSections, err := labSectionService.GetPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labSectionService.Count(c.RequestCtx(), filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
		}

		type labSectionResponse struct {
			ID                 string     `json:"id"`
			Name               string     `json:"name"`
			ReadonlyAt           *time.Time `json:"readonly_at,omitempty"`
			Status             string     `json:"status,omitempty"`
			TotalMaterials     int        `json:"total_materials"`
			CompletedMaterials int        `json:"completed_materials"`
			StudentStatus      string     `json:"student_status"`
		}

		responseSections := make([]labSectionResponse, len(labSections))
		for i, section := range labSections {
			lab, err := labService.GetByID(c.RequestCtx(), section.LabID)
			if err != nil {
				return err
			}

			totalMaterials, err := labMaterialService.Count(c.RequestCtx(), map[string]string{"lab_id__is": section.LabID})
			if err != nil {
				return err
			}

			allMaterials, err := labMaterialService.GetByLabID(c.RequestCtx(), section.LabID)
			if err != nil {
				return err
			}

			completedMaterials := 0
			hasAnySubmission := false
			for _, mat := range allMaterials {
				status := submissionService.GetMaterialStudentStatus(c.RequestCtx(), user.ID, mat.MaterialID, section.LabID, section.SectionID)
				if status != "not_started" {
					hasAnySubmission = true
					if status == "passed" {
						completedMaterials++
					}
				}
			}

			studentStatus := "not_started"
			if totalMaterials > 0 && completedMaterials == totalMaterials {
				studentStatus = "passed"
			} else if completedMaterials > 0 {
				studentStatus = "in_progress"
			} else if hasAnySubmission {
				studentStatus = "not_passed"
			}

			responseSections[i] = labSectionResponse{
				ID:                 section.LabID,
				Name:               lab.DisplayName,
				ReadonlyAt:         section.ReadonlyAt,
				Status:             section.EffectiveStatus(),
				TotalMaterials:     totalMaterials,
				CompletedMaterials: completedMaterials,
				StudentStatus:      studentStatus,
			}
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
				"total_rows": count,
			},
			"data": responseSections,
		})
	})

	coreSectionRouter.Delete("/:sectionID/unenroll", middlewares.Permission(permService).ForSection("sectionID").CanView(), func(c fiber.Ctx) error {
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
