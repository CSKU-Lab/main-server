package routes

import (
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/gofiber/fiber/v3"
)

const maxSearchQueryLen = 100

func NewCoreSearchRoutes(router fiber.Router, searchService services.SearchService) {
	router.Get("/search", func(c fiber.Ctx) error {
		q := c.Query("q", "")
		if q == "" {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "q is required",
			})
		}
		if len(q) > maxSearchQueryLen {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "q exceeds maximum length",
			})
		}

		limitStr := c.Query("limit", "5")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 20 {
			limit = 5
		}

		user := c.Locals("user").(*models.User)

		result, err := searchService.SearchForStudent(c.RequestCtx(), user.ID, q, limit)
		if err != nil {
			return err
		}

		return c.JSON(result)
	})
}
