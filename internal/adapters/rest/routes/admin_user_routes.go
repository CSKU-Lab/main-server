package routes

import (
	"errors"
	"math"
	"strconv"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/gofiber/fiber/v2"
)

type userRoutes struct {
	router      fiber.Router
	userService services.UserService
}

func NewAdminUserRoutes(router fiber.Router, userService services.UserService) {
	adminUserRouter := router.Group("/users")

	adminUserRouter.Get("/", func(c *fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("pageSize", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "created_at")
		sortOrder := c.Query("sort_order", "desc")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid page")
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid page size")
		}

		users, err := userService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting users")
		}

		count, err := userService.Count(c.Context(), search)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting users count")
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"users": users,
		})
	})

	adminUserRouter.Post("/oauth", func(c *fiber.Ctx) error {
		var userRequest requests.User

		err := c.BodyParser(&userRequest)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Error parsing request")
		}

		user, err := userService.Create(c.Context(), &userRequest)
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error creating user")
		}

		return c.Status(fiber.StatusCreated).JSON(user)
	})

	adminUserRouter.Post("/credential", func(c *fiber.Ctx) error {
		var userRequest requests.CredentialUser

		err := c.BodyParser(&userRequest)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Error parsing request")
		}

		user, err := userService.Create(c.Context(), &requests.User{
			Username:    userRequest.Username,
			DisplayName: userRequest.DisplayName,
			Roles:       userRequest.Roles,
		})
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error creating user")
		}

		err = userService.SetPassword(c.Context(), user.ID, userRequest.Password)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error creating user")
		}

		return c.Status(fiber.StatusCreated).JSON(user)
	})

	router.Get("/users/:userID", func(c *fiber.Ctx) error {
		userID := c.Params("userID")

		user, err := userService.GetByID(c.Context(), userID)
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting user")
		}

		return c.JSON(user)
	})

	adminUserRouter.Patch("/:userID", func(c *fiber.Ctx) error {
		var updateUser requests.User
		err := c.BodyParser(&updateUser)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Invalid request body")
		}

		user, err := userService.Update(c.Context(), c.Params("userID"), &updateUser)
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error updating user")
		}

		return c.JSON(user)
	})

	adminUserRouter.Delete("/:userID", func(c *fiber.Ctx) error {
		err := userService.Delete(c.Context(), c.Params("userID"))
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error deleting user")
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
