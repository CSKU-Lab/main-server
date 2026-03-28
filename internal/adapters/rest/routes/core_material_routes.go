package routes

import (
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

func NewCoreMaterialSubmissionRoutes(
	router fiber.Router,
	materialService services.MaterialService,
	submissionService services.SubmissionService,
	submissionRepo repositories.SubmissionRepository,
	labSectionService services.LabSectionService,
) {
	materialRouter := router.Group("/materials")

	// GET /api/v1/core/materials/:materialID - Must be enrolled student
	materialRouter.Get("/:materialID",
		middlewares.PermissionMiddleware(func(user *models.User, c fiber.Ctx) error {
			sectionID := c.Query("section_id", "")
			if sectionID == "" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusBadRequest,
					Message:    "section_id query parameter is required",
				})
			}
			_, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), c.Query("lab_id", ""), sectionID)
			if err != nil {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusForbidden,
					Message:    "You do not have access to this section",
				})
			}
			return nil
		}),
		func(c fiber.Ctx) error {
			materialID := c.Params("materialID")
			user := c.Locals("user").(*models.User)
			sectionID := c.Query("section_id", "")
			labID := c.Query("lab_id", "")

			labSec, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, sectionID)
			if err != nil {
				return err
			}

			result, err := materialService.GetMaterialWithLatestSubmissionStatus(c.RequestCtx(), user.ID, materialID, labSec.LabID, labSec.SectionID)
			if err != nil {
				return err
			}

			return c.JSON(result)
		},
	)

	// GET /api/v1/core/materials/:materialID/submissions - Must be enrolled student (own submissions)
	materialRouter.Get("/:materialID/submissions",
		middlewares.PermissionMiddleware(func(user *models.User, c fiber.Ctx) error {
			sectionID := c.Query("section_id", "")
			if sectionID == "" {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusBadRequest,
					Message:    "section_id query parameter is required",
				})
			}
			_, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), c.Query("lab_id", ""), sectionID)
			if err != nil {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusForbidden,
					Message:    "You do not have access to this section",
				})
			}
			return nil
		}),
		func(c fiber.Ctx) error {
			materialID := c.Params("materialID")
			user := c.Locals("user").(*models.User)

			sectionID := c.Query("section_id", "")
			labID := c.Query("lab_id", "")

			pageQuery := c.Query("page", "1")
			pageSizeQuery := c.Query("page_size", "20")
			sortOrder := c.Query("sort_order", "desc")

			page, err := strconv.Atoi(pageQuery)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
			}

			pageSize, err := strconv.Atoi(pageSizeQuery)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
			}

			if sortOrder != "asc" && sortOrder != "desc" {
				sortOrder = "desc"
			}

			labSec, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, sectionID)
			if err != nil {
				return err
			}

			submissions, count, err := submissionService.GetUserSubmissionsWithMaterial(c.RequestCtx(), user.ID, materialID, labSec.LabID, labSec.SectionID, page, pageSize, sortOrder)
			if err != nil {
				return err
			}

			return c.JSON(fiber.Map{
				"pagination": fiber.Map{
					"page":       page,
					"total_page": int(math.Ceil(float64(count) / float64(pageSize))),
					"total_rows": count,
				},
				"data": submissions,
			})
		},
	)
}
