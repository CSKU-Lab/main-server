package routes

import (
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

func NewCoreMaterialSubmissionRoutes(
	router fiber.Router,
	materialService services.MaterialService,
	submissionService services.SubmissionService,
	labSectionService services.LabSectionService,
	permService permission.Service,
) {
	materialRouter := router.Group("/materials")

	materialRouter.Get("/:materialID", middlewares.Permission(permService).ForSection("section_id").FromQuery().CanView(), func(c fiber.Ctx) error {
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
	})

	materialRouter.Get("/:materialID/submissions", middlewares.Permission(permService).ForSection("section_id").FromQuery().CanView(), func(c fiber.Ctx) error {
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
	})
}
