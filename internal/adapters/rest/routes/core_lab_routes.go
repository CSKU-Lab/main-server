package routes

import (
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

func NewCoreLabRoute(router fiber.Router, sectionService services.SectionService, labSectionService services.LabSectionService, labService services.LabService, sectionStudentService services.SectionStudentService, labMaterialService services.LabMaterialService, permService permission.Service) {
	coreLabRoute := router.Group("/labs")

	coreLabRoute.Post("/:labID", middlewares.ValidateMiddleware[requests.GetSection](), middlewares.Permission(permService).ForSection("section_id").CanView(), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.GetSection)
		labID := c.Params("labID")
		user := c.Locals("user").(*models.User)

		secStudent, err := sectionStudentService.GetBySectionAndStudentID(c.RequestCtx(), req.SectionID, user.ID)
		if err != nil {
			return err
		}
		labSection, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, secStudent.SectionID)
		if err != nil {
			return err
		}
		lab, err := labService.GetByID(c.RequestCtx(), labSection.LabID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(lab)
	})

	coreLabRoute.Get("/:labID/materials", middlewares.Permission(permService).ForSection("section_id").CanView(), func(c fiber.Ctx) error {
		labID := c.Params("labID")
		sectionID := c.Query("section_id", "")
		user := c.Locals("user").(*models.User)

		// Check if user is a section student
		secStudent, err := sectionStudentService.GetBySectionAndStudentID(c.RequestCtx(), sectionID, user.ID)
		if err != nil {
			// If not a student, check if user is instructor or admin
			isInstructor, permErr := permService.IsSectionInstructor(c.Context(), user.ID, sectionID)
			if permErr != nil {
				return permErr
			}

			isAdmin, permErr := permService.IsAdmin(c.Context(), user.ID)
			if permErr != nil {
				return permErr
			}

			// If neither instructor nor admin, return the original error
			if !isInstructor && !isAdmin {
				return err
			}

			// For instructors/admins, get the lab section directly
			labSection, permErr := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, sectionID)
			if permErr != nil {
				return permErr
			}

			pageQuery := c.Query("page", "1")
			pageSizeQuery := c.Query("page_size", "10")
			sortBy := c.Query("sort_by", "created_at")
			sortOrder := c.Query("sort_order", "asc")

			filterParams := make(map[string]string)
			for key, value := range c.Queries() {
				if strings.Contains(key, "__") {
					filterParams[key] = value
				}
			}

			page, convErr := strconv.Atoi(pageQuery)
			if convErr != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
			}

			pageSize, convErr := strconv.Atoi(pageSizeQuery)
			if convErr != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
			}

			filterParams["lab_id__is"] = labSection.LabID

			materials, permErr := labMaterialService.GetPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
			if permErr != nil {
				return permErr
			}

			count, permErr := labMaterialService.Count(c.RequestCtx(), filterParams)
			if permErr != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
			}

			return c.JSON(fiber.Map{
				"pagination": fiber.Map{
					"page":       page,
					"total_page": math.Ceil(float64(count/pageSize) + 1),
					"total_rows": count,
				},
				"data": materials,
			})
		}

		// Student path - continue with existing logic
		labSection, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, secStudent.SectionID)
		if err != nil {
			return err
		}

		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		sortBy := c.Query("sort_by", "created_at")
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

		filterParams["lab_id__is"] = labSection.LabID

		materials, err := labMaterialService.GetPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labMaterialService.Count(c.RequestCtx(), filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": materials,
		})
	})
}

// fiber:context-methods migrated
