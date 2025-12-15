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
	"github.com/gofiber/fiber/v2"
)

func NewCMSLabRoutes(router fiber.Router, labService services.LabService, labSectionService services.LabSectionService, labMaterialService services.LabMaterialService) {
	labRouter := router.Group("/labs", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	labRouter.Get("/:labID", func(c *fiber.Ctx) error {
		labID := c.Params("labID")
		lab, err := labService.GetByID(c.Context(), labID)
		if err != nil {
			return err
		}
		return c.JSON(lab)
	})

	labRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateLab](), func(c *fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateLab)

		user := c.Locals("user").(*models.User)

		labID, err := labService.Create(c.Context(), req, user.ID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": labID,
		})
	})

	labRouter.Get("/", func(c *fiber.Ctx) error {
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

		labs, err := labService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labService.Count(c.Context(), search, filterParams)
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

	labRouter.Patch("/:labID", middlewares.ValidateMiddleware[requests.BaseUpdateLab](), func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		req := c.Locals("body").(*requests.BaseUpdateLab)

		return labService.UpdateByID(c.Context(), labID, user.ID, req)
	})

	labRouter.Delete("/:labID", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		return labService.DeleteByID(c.Context(), labID, user.ID)
	})

	labRouter.Post("/set-section", middlewares.ValidateMiddleware[requests.SetLabSection](), func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		req := c.Locals("body").(*requests.SetLabSection)
		err := labSectionService.Create(c.Context(), req, user.ID)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	labRouter.Get("/:labID/sections", func(c *fiber.Ctx) error {
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

		sections, err := labSectionService.GetPagination(c.Context(), page, pageSize, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labSectionService.Count(c.Context(), filterParams)
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

	labRouter.Patch("/:labID/sections/:sectionID", middlewares.ValidateMiddleware[requests.UpdateLabSection](), func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		sectionID := c.Params("sectionID")
		req := c.Locals("body").(*requests.UpdateLabSection)

		return labSectionService.UpdateByID(c.Context(), user.ID, labID, sectionID, req)
	})

	labRouter.Delete("/:labID/sections/:sectionID", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		sectionID := c.Params("sectionID")
		return labSectionService.DeleteByID(c.Context(), labID, sectionID, user.ID)
	})

	labRouter.Post("/set-material", middlewares.ValidateMiddleware[requests.SetLabMaterial](), func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		req := c.Locals("body").(*requests.SetLabMaterial)
		err := labMaterialService.Create(c.Context(), req, user.ID)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusCreated)
	})
}
