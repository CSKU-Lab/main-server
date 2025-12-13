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

func NewCMSLabRoutes(router fiber.Router, labService services.LabService) {
	labRouter := router.Group("/labs")

	labRouter.Get("/:labID", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c *fiber.Ctx) error {
		labID := c.Params("labID")
		lab, err := labService.GetByID(c.Context(), labID)
		if err != nil {
			return err
		}
		return c.JSON(lab)
	})

	labRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateLab](), middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c *fiber.Ctx) error {
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

	labRouter.Get("/", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c *fiber.Ctx) error {
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

	labRouter.Patch("/:labID", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.ValidateMiddleware[requests.BaseUpdateLab](), func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		req := c.Locals("body").(*requests.BaseUpdateLab)

		return labService.UpdateByID(c.Context(), labID, user.ID, req)
	})

	labRouter.Delete("/:labID", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		labID := c.Params("labID")
		return labService.DeleteByID(c.Context(), labID, user.ID)
	})
}
