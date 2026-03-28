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

func NewCMSMaterialRoutes(router fiber.Router, materialService services.MaterialService, materialAssetService services.MaterialAssetService, submissionService services.SubmissionService) {
	// Material routes nested under sections
	// Routes are mounted at /sections/:id/materials
	materialRouter := router.Group("/sections/:id/materials")

	// GET /api/v1/cms/sections/:id/materials
	// Permission: IsAdmin OR IsSectionInstructor OR IsSectionStudent
	// Students can view materials in their sections
	materialRouter.Get("/",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
				permission.IsSectionStudent("id"),
			),
		),
		func(c fiber.Ctx) error {
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

			mats, err := materialService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, filterParams)
			if err != nil {
				return err
			}

			count, err := materialService.Count(c.RequestCtx(), search, filterParams)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting materials count"})
			}

			return c.JSON(fiber.Map{
				"pagination": fiber.Map{
					"page":       page,
					"total_page": math.Ceil(float64(count/pageSize) + 1),
					"total_rows": count,
				},
				"data": mats,
			})
		})

	// POST /api/v1/cms/sections/:id/materials
	// Permission: IsAdmin OR IsSectionInstructor
	// Only instructors can create materials
	materialRouter.Post("/",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
			),
		),
		middlewares.ValidateMiddleware[requests.CreateMaterial](),
		func(c fiber.Ctx) error {
			req := c.Locals("body").(*requests.CreateMaterial)
			user := c.Locals("user").(*models.User)

			matID, err := materialService.Create(c.RequestCtx(), user.ID, req)
			if err != nil {
				return err
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"id": matID,
			})
		})

	// GET /api/v1/cms/sections/:id/materials/:materialId
	// Permission: IsAdmin OR IsSectionInstructor OR IsSectionStudent
	// Students can view materials in their sections
	materialRouter.Get("/:materialId",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
				permission.IsSectionStudent("id"),
			),
		),
		func(c fiber.Ctx) error {
			id := c.Params("materialId")
			material, err := materialService.GetByID(c.RequestCtx(), id)
			if err != nil {
				return err
			}
			return c.JSON(material)
		})

	// PATCH /api/v1/cms/sections/:id/materials/:materialId
	// Permission: IsAdmin OR IsSectionInstructor
	// Only instructors can update materials
	materialRouter.Patch("/:materialId",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
			),
		),
		middlewares.ValidateMiddleware[requests.BaseUpdateMaterial](),
		func(c fiber.Ctx) error {
			user := c.Locals("user").(*models.User)
			id := c.Params("materialId")
			req := c.Locals("body").(*requests.BaseUpdateMaterial)
			rawReq := c.Body()

			return materialService.UpdateByID(c.RequestCtx(), id, req, rawReq, user.ID)
		})

	// DELETE /api/v1/cms/sections/:id/materials/:materialId
	// Permission: IsAdmin OR IsSectionInstructor
	// Only instructors can delete materials
	materialRouter.Delete("/:materialId",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
			),
		),
		func(c fiber.Ctx) error {
			user := c.Locals("user").(*models.User)
			id := c.Params("materialId")
			return materialService.DeleteByID(c.RequestCtx(), id, user.ID)
		})

	// POST /api/v1/cms/sections/:id/materials/:materialId/assets
	// Permission: IsAdmin OR IsSectionInstructor
	// Only instructors can upload assets
	materialRouter.Post("/:materialId/assets",
		middlewares.RequirePermission(
			permission.Or(
				permission.IsAdmin,
				permission.IsSectionInstructor("id"),
			),
		),
		func(c fiber.Ctx) error {
			filePayload, err := c.FormFile("file")
			if err != nil {
				return err
			}

			file, err := filePayload.Open()
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to open image file"})
			}

			defer file.Close()

			imageFile := &requests.File{
				Name:        filePayload.Filename,
				Size:        filePayload.Size,
				Reader:      file,
				ContentType: filePayload.Header.Get("Content-Type"),
			}

			id := c.Params("materialId")

			fileURL, err := materialAssetService.UploadFile(c.RequestCtx(), id, imageFile)
			if err != nil {
				return err
			}

			return c.JSON(fiber.Map{
				"url": fileURL,
			})
		})
}

// fiber:context-methods migrated
