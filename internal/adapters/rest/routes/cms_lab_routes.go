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
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCMSLabRoutes(router fiber.Router, labService services.LabService, labSectionService services.LabSectionService, labMaterialService services.LabMaterialService, sectionService services.SectionService, submissionService services.SubmissionService) {
	labRouter := router.Group("/labs", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	labRouter.Get("/:labID", func(c fiber.Ctx) error {
		labID := c.Params("labID")
		lab, err := labService.GetByID(c.RequestCtx(), labID)
		if err != nil {
			return err
		}
		return c.JSON(lab)
	})

	labRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateLab](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateLab)

		user := c.Locals("user").(*models.User)

		labID, err := labService.Create(c.RequestCtx(), req, user.ID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": labID,
		})
	})

	labRouter.Get("/", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "display_name")
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

	labRouter.Patch("/:labID", middlewares.ValidateMiddleware[requests.BaseUpdateLab](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		req := c.Locals("body").(*requests.BaseUpdateLab)

		return labService.UpdateByID(c.RequestCtx(), labID, user.ID, req)
	})

	labRouter.Delete("/:labID", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		return labService.DeleteByID(c.RequestCtx(), labID, user.ID)
	})

	labRouter.Get("/:labID/sections", func(c fiber.Ctx) error {
		labID := c.Params("labID")

		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
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

		filterParams["lab_id__is"] = labID

		sections, err := labSectionService.GetPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labSectionService.Count(c.RequestCtx(), filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": sections,
		})
	})

	labRouter.Post("/:labID/materials", middlewares.ValidateMiddleware[requests.SetLabMaterial](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		req := c.Locals("body").(*requests.SetLabMaterial)
		err := labMaterialService.Create(c.RequestCtx(), req, user.ID, labID)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	labRouter.Get("/:labID/materials/all", func(c fiber.Ctx) error {
		labID := c.Params("labID")
		labMaterials, err := labMaterialService.GetByLabID(c.RequestCtx(), labID)
		if err != nil {
			return err
		}
		return c.JSON(labMaterials)
	})

	labRouter.Get("/:labID/materials", func(c fiber.Ctx) error {
		labID := c.Params("labID")

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

		filterParams["lab_id__is"] = labID

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

	labRouter.Post("/:labID/materials/delete", middlewares.ValidateMiddleware[requests.DeleteLabMaterial](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		req := c.Locals("body").(*requests.DeleteLabMaterial)
		return labMaterialService.Delete(c.RequestCtx(), labID, user.ID, req)
	})

	// Section-based lab routes with permission guards
	// These routes are nested under /sections/:id/labs
	sectionLabRouter := router.Group("/sections/:id/labs")

	// GET /api/v1/cms/sections/:id/labs - List labs in a section
	// Permission: IsAdmin OR IsSectionInstructor OR IsSectionStudent
	sectionLabRouter.Get("/",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
				permission.IsSectionStudent("id"),
			),
		),
		func(c fiber.Ctx) error {
			sectionID := c.Params("id")

			pageQuery := c.Query("page", "1")
			pageSizeQuery := c.Query("page_size", "10")
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

			filterParams["section_id__is"] = sectionID

			labSections, err := labSectionService.GetPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
			if err != nil {
				return err
			}

			count, err := labSectionService.Count(c.RequestCtx(), filterParams)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
			}

			type labSectionResponse struct {
				LabID    string     `json:"lab_id"`
				Position int        `json:"position"`
				Status   string     `json:"status"`
				OpenedAt *time.Time `json:"opened_at"`
				ClosedAt *time.Time `json:"closed_at"`
				LabName  string     `json:"lab_name"`
			}

			responseSections := make([]labSectionResponse, len(labSections))
			for i, section := range labSections {
				lab, err := labService.GetByID(c.RequestCtx(), section.LabID)
				if err != nil {
					return err
				}

				responseSections[i] = labSectionResponse{
					LabID:    section.LabID,
					Position: section.Position,
					Status:   section.Status,
					OpenedAt: section.OpenedAt,
					ClosedAt: section.ClosedAt,
					LabName:  lab.DisplayName,
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
		},
	)

	// POST /api/v1/cms/sections/:id/labs - Create lab in a section
	// Permission: IsAdmin OR IsSectionInstructor
	sectionLabRouter.Post("/",
		middlewares.ValidateMiddleware[requests.SetLabSection](),
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
			),
		),
		func(c fiber.Ctx) error {
			user := c.Locals("user").(*models.User)
			sectionID := c.Params("id")
			req := c.Locals("body").(*requests.SetLabSection)

			ctx := c.Context()
			err := labSectionService.Create(ctx, req, user.ID, sectionID)
			if err != nil {
				return err
			}

			return c.SendStatus(fiber.StatusCreated)
		},
	)

	// GET /api/v1/cms/sections/:id/labs/:labId - Get lab details in a section
	// Permission: IsAdmin OR IsSectionInstructor OR IsSectionStudent
	sectionLabRouter.Get("/:labId",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
				permission.IsSectionStudent("id"),
			),
		),
		func(c fiber.Ctx) error {
			sectionID := c.Params("id")
			labID := c.Params("labId")

			// Get lab section details
			labSection, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, sectionID)
			if err != nil {
				return err
			}

			// Get lab name
			lab, err := labService.GetByID(c.RequestCtx(), labID)
			if err != nil {
				return err
			}

			// Get total students
			students, err := sectionService.GetStudentsBySectionID(c.RequestCtx(), sectionID)
			if err != nil {
				return err
			}
			totalStudents := len(students)

			// Get completed students (passed all materials)
			completedStudents, err := submissionService.CountCompletedStudentsByLabAndSection(c.RequestCtx(), labID, sectionID)
			if err != nil {
				return err
			}

			type labDetailResponse struct {
				LabName           string     `json:"lab_name"`
				Status            string     `json:"status"`
				OpenedAt          *time.Time `json:"opened_at"`
				ClosedAt          *time.Time `json:"closed_at"`
				CompletedStudents int        `json:"completed_students"`
				TotalStudents     int        `json:"total_students"`
			}

			return c.JSON(labDetailResponse{
				LabName:           lab.DisplayName,
				Status:            labSection.Status,
				OpenedAt:          labSection.OpenedAt,
				ClosedAt:          labSection.ClosedAt,
				CompletedStudents: completedStudents,
				TotalStudents:     totalStudents,
			})
		},
	)

	// PATCH /api/v1/cms/sections/:id/labs/:labId - Update lab in a section
	// Permission: IsAdmin OR IsSectionInstructor
	sectionLabRouter.Patch("/:labId",
		middlewares.ValidateMiddleware[requests.UpdateLabSectionStatus](),
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
			),
		),
		func(c fiber.Ctx) error {
			user := c.Locals("user").(*models.User)
			sectionID := c.Params("id")
			labID := c.Params("labId")
			req := c.Locals("body").(*requests.UpdateLabSectionStatus)

			err := labSectionService.UpdateStatus(c.RequestCtx(), user.ID, sectionID, labID, req)
			if err != nil {
				return err
			}

			return c.SendStatus(fiber.StatusAccepted)
		},
	)

	// DELETE /api/v1/cms/sections/:id/labs/:labId - Delete lab from a section
	// Permission: IsAdmin OR IsSectionInstructor
	sectionLabRouter.Delete("/:labId",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
			),
		),
		func(c fiber.Ctx) error {
			user := c.Locals("user").(*models.User)
			sectionID := c.Params("id")
			labID := c.Params("labId")

			req := &requests.DeleteLabSection{
				LabIDs: []string{labID},
			}

			ctx := c.Context()
			return labSectionService.Delete(ctx, sectionID, user.ID, req)
		},
	)
}

// fiber:context-methods migrated
