package routes

import (
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewAdminUserGroupRoutes(router fiber.Router, userGroupService services.UserGroupService) {
	adminUserGroupRoutes := router.Group("/user-groups")

	adminUserGroupRoutes.Post("/", func(c *fiber.Ctx) error {
		var req requests.UserGroup
		err := c.BodyParser(&req)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Cannot parse the body",
			})
		}

		_, err = userGroupService.Create(c.Context(), req.Name)
		return err
	})

	adminUserGroupRoutes.Get("/", func(c *fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "desc")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid page",
			})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid page size",
			})
		}

		userGroups, err := userGroupService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder)
		if err != nil {
			return err
		}

		count, err := userGroupService.Count(c.Context(), search)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": userGroups,
		})
	})

	adminUserGroupRoutes.Patch("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req requests.UserGroup
		err := c.BodyParser(&req)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Cannot parse the body",
			})
		}

		err = userGroupService.Update(c.Context(), id, req.Name)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusAccepted)
	})

	adminUserGroupRoutes.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return userGroupService.Delete(c.Context(), id)
	})

}
