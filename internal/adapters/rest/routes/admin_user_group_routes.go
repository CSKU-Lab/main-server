package routes

import (
	"math"
	"strconv"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/responses"
	"github.com/gofiber/fiber/v2"
)

func NewAdminUserGroupRoutes(router fiber.Router, userGroupService services.UserGroupService) {
	adminUserGroupRoutes := router.Group("/user-groups")

	adminUserGroupRoutes.Post("/", func(c *fiber.Ctx) error {
		var req requests.UserGroup
		err := c.BodyParser(&req)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot parse the body")
		}

		return userGroupService.Create(c.Context(), req.Name)
	})

	adminUserGroupRoutes.Get("/", func(c *fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "desc")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid page")
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid page size")
		}

		userGroups, err := userGroupService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder)
		if err != nil {
			return err
		}

		userGroupsResponse := make([]responses.UserGroup, len(userGroups))
		for i, userGroup := range userGroups {
			userGroupsResponse[i] = responses.UserGroup{
				ID:   userGroup.ID,
				Name: userGroup.Name,
			}
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
			"data": userGroupsResponse,
		})
	})

	adminUserGroupRoutes.Patch("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req requests.UserGroup
		err := c.BodyParser(&req)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot parse the body")
		}

		updatedGroup, err := userGroupService.Update(c.Context(), id, req.Name)
		if err != nil {
			return err
		}

		return c.JSON(&models.UserGroup{
			ID:   updatedGroup.ID,
			Name: updatedGroup.Name,
		})

	})

	adminUserGroupRoutes.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return userGroupService.Delete(c.Context(), id)
	})

}
