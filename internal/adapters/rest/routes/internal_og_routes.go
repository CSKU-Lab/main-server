package routes

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/queue"
	"github.com/gofiber/fiber/v3"
)

func NewInternalOGRoutes(router fiber.Router, internalToken string, courseService services.CourseService, sectionService services.SectionService, materialService services.MaterialService, q queue.Queue) {
	router.Post("/internal/og/resync", func(c fiber.Ctx) error {
		if internalToken == "" || c.Get("Authorization") != "Bearer "+internalToken {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		go resyncAllOGImages(courseService, sectionService, materialService, q)

		return c.JSON(fiber.Map{"status": "resync started"})
	})
}

func resyncAllOGImages(courseService services.CourseService, sectionService services.SectionService, materialService services.MaterialService, q queue.Queue) {
	ctx := context.Background()

	courses, err := courseService.GetPagination(ctx, 1, 10000, "", "name", "asc", "active", "")
	if err == nil {
		for _, course := range courses {
			publishOGEvent(q, OGImageEvent{Type: "course", ID: course.ID, Title: course.Name})

			materials, err := materialService.GetPagination(ctx, course.ID, "", []models.Role{models.ADMIN}, 1, 10000, "", "name", "asc", nil)
			if err == nil {
				for _, mat := range materials {
					tag := mat.Type
					publishOGEvent(q, OGImageEvent{Type: "material", ID: mat.ID, Title: mat.Name, Tag: &tag})
				}
			}
		}
	}

	sections, err := sectionService.GetSectionsPagination(ctx, 1, 10000, "id", "asc", nil)
	if err == nil {
		for _, sec := range sections {
			tag := sec.Semester.Name
			publishOGEvent(q, OGImageEvent{Type: "section", ID: sec.ID, Title: sec.Name, Tag: &tag})
		}
	}
}
