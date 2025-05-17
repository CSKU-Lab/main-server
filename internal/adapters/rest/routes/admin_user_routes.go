package routes

import (
	"errors"
	"log"
	"math"
	"strconv"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
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
		pageSizeQuery := c.Query("page_size", "10")
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
			return err
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
		var userRequest requests.CreateUser

		err := c.BodyParser(&userRequest)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Error parsing request")
		}

		user, err := userService.Create(c.Context(), models.UserTypeOauth, &userRequest)
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
		var userRequest requests.CreateCredentialUser

		err := c.BodyParser(&userRequest)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Error parsing request")
		}

		user, err := userService.Create(c.Context(), models.UserTypeCredential, &requests.CreateUser{
			BaseUser: requests.BaseUser{
				Username:    userRequest.Username,
				DisplayName: userRequest.DisplayName,
				Roles:       userRequest.Roles,
			},
			Email: nil,
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

	// TODO: Add validation for user type
	adminUserRouter.Patch("/:userID", func(c *fiber.Ctx) error {
		var updateUser requests.UpdateUser
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

	adminUserRouter.Post("/deleteMany", func(c *fiber.Ctx) error {
		var deleteManyUser requests.DeleteManyUser

		err := c.BodyParser(&deleteManyUser)
		if err != nil {
			return cserrors.New(cserrors.BAD_REQUEST, "Error parsing request")
		}

		err = userService.DeleteMany(c.Context(), deleteManyUser.IDs)
		if err != nil {
			log.Println(err)
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error deleting users")
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
