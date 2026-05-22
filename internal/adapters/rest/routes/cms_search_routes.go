package routes

import (
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v3"
)

func NewCMSSearchRoutes(router fiber.Router, searchService services.SearchService) {
	router.Get("/search", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), func(c fiber.Ctx) error {
		q := c.Query("q", "")
		if q == "" {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "q is required",
			})
		}

		limitStr := c.Query("limit", "5")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 20 {
			limit = 5
		}

		result, err := searchService.Search(c.RequestCtx(), q, limit)
		if err != nil {
			return err
		}

		return c.JSON(result)
	})
}
