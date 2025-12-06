package routes

import (
	"errors"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/gofiber/fiber/v2"
)

func NewCMSCourseRoutes(router fiber.Router, service services.SectionService) {
	courseRouter := router.Group("/course")

	courseRouter.Get("/:courseID/sections", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
		models.STUDENT,
	}), func(c *fiber.Ctx) error {
		courseID := c.Params("courseID")
		sections, err := service.GetByCourseID(c.Context(), courseID)
		if err != nil {
			var csErr *cserrors.Error
			if errors.As(err, &csErr) {
				return err
			}
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Error getting sections by course ID",
			})
		}

		return c.JSON(sections)
	})
}
