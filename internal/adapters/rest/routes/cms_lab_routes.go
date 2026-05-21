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
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCMSLabRoutes(router fiber.Router, labService services.LabService, labSectionService services.LabSectionService, labMaterialService services.LabMaterialService) {
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
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
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
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
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
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
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

	labRouter.Patch("/:labID/materials/:materialID/position", middlewares.ValidateMiddleware[requests.UpdateLabMaterialPosition](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		materialID := c.Params("materialID")
		req := c.Locals("body").(*requests.UpdateLabMaterialPosition)
		return labMaterialService.UpdatePosition(c.RequestCtx(), labID, materialID, user.ID, req.Position)
	})
}

// fiber:context-methods migrated
