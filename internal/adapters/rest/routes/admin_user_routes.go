package routes

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewAdminUserRoutes(router fiber.Router, userService services.UserService) {
	adminUserRouter := router.Group("/users", middlewares.RequireAdmin())

	adminUserRouter.Get("/", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "created_at")
		sortOrder := c.Query("sort_order", "desc")

		filterParams := make(map[string]string)
		for key, value := range c.Queries() {
			if strings.Contains(key, "__") {
				filterParams[key] = value
			}
		}

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

		users, err := userService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := userService.Count(c.RequestCtx(), search, filterParams)
		if err != nil {
			return cserrors.New(
				&cserrors.Option{
					HttpStatus: http.StatusInternalServerError,
					Message:    "Error getting users count",
				})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
				"total_rows": count,
			},
			"data": users,
		})
	})

	adminUserRouter.Post("/", middlewares.ValidateMiddleware[requests.CreateMultiTypeUser](), func(c fiber.Ctx) error {
		userRequest := c.Locals("body").(*requests.CreateMultiTypeUser)

		err := userService.Create(c.RequestCtx(), userRequest)
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error creating user",
			})
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	adminUserRouter.Post("/import", middlewares.ValidateMiddleware[requests.CreateManyUsers](), func(c fiber.Ctx) error {
		createManyUsers := c.Locals("body").(*requests.CreateManyUsers)

		err := userService.UpsertMany(c.RequestCtx(), createManyUsers)
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error importing users",
			})
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	adminUserRouter.Get("/:userID", func(c fiber.Ctx) error {
		userID := c.Params("userID")

		user, err := userService.GetByID(c.RequestCtx(), userID)
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error getting user",
			})
		}

		return c.JSON(user)
	})

	adminUserRouter.Patch("/:userID", middlewares.ValidateMiddleware[requests.UpdateUser](), func(c fiber.Ctx) error {
		updateUser := c.Locals("body").(*requests.UpdateUser)

		err := userService.Update(c.RequestCtx(), c.Params("userID"), updateUser)
		if err != nil {
			var e *cserrors.Error
			if errors.As(err, &e) {
				return e
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error updating user",
			})
		}

		return c.SendStatus(fiber.StatusAccepted)
	})

	adminUserRouter.Delete("/:userID", func(c fiber.Ctx) error {
		err := userService.Delete(c.RequestCtx(), c.Params("userID"))
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error deleting uesr",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	adminUserRouter.Post("/deleteMany", middlewares.ValidateMiddleware[requests.DeleteManyUser](), func(c fiber.Ctx) error {
		deleteManyUser := c.Locals("body").(*requests.DeleteManyUser)

		err := userService.DeleteMany(c.RequestCtx(), deleteManyUser.IDs)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error deleting uesr",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}

// fiber:context-methods migrated
