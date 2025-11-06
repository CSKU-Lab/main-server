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

func NewCMSMaterialRoutes(router fiber.Router, materialService services.MaterialService) {
	materialRouter := router.Group("/materials")

	materialRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateMaterial](), func(c *fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateMaterial)
		user := c.Locals("user").(*models.User)

		return materialService.Create(c.Context(), user.ID, req)
	})

	materialRouter.Get("/", func(c *fiber.Ctx) error {
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

		sems, err := materialService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := materialService.Count(c.Context(), search, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semesters count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": sems,
		})
	})

	materialRouter.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		material, err := materialService.GetByID(c.Context(), id)
		if err != nil {
			return err
		}
		return c.JSON(material)
	})

	materialRouter.Patch("/:id", middlewares.ValidateMiddleware[requests.UpdateMaterial](), func(c *fiber.Ctx) error {
		id := c.Params("id")
		req := c.Locals("body").(*requests.UpdateMaterial)

		return materialService.UpdateByID(c.Context(), id, req)
	})

	materialRouter.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return materialService.DeleteByID(c.Context(), id)
	})
}
