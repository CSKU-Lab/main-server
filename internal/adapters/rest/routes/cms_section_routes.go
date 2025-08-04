package routes

import (
	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/gofiber/fiber/v2"
)

type cmsSectionRoutes struct {
	router         fiber.Router
	sectionService services.SectionService
}

func NewCmsSectionRoutes(router fiber.Router, sectionService services.SectionService) {
	cmsSectionRouter := router.Group("/sections")

	cmsSectionRouter.Post("/", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		courseID := c.FormValue("course_id")
		semesterID := c.FormValue("semester_id")
		startedAt := c.FormValue("started_at")
		endedAt := c.FormValue("ended_at")

		req := requests.Section{
			Name:      name,
			StartedAt: startedAt,
			EndedAt:   endedAt,
			Image:     nil,
		}

		image, err := c.FormFile("image")
		if err == nil {
			file, err := image.Open()
			if err != nil {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to open image file")
			}

			defer file.Close()

			imageFile := &requests.File{
				Name:        image.Filename,
				Size:        image.Size,
				Reader:      file,
				ContentType: image.Header.Get("Content-Type"),
			}

			req.Image = imageFile
		}

		createdSection, err := sectionService.Create(c.Context(), &requests.CreateSection{
			Section:    req,
			CourseID:   courseID,
			SemesterID: semesterID,
		})
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":         createdSection.ID,
			"name":       createdSection.Name,
			"image":      createdSection.Image,
			"started_at": createdSection.StartedAt,
			"ended_at":   createdSection.EndedAt,
		})
	})

	cmsSectionRouter.Patch("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		name := c.FormValue("name")
		startedAt := c.FormValue("started_at")
		endedAt := c.FormValue("ended_at")

		req := &requests.Section{
			Name:      name,
			StartedAt: startedAt,
			EndedAt:   endedAt,
			Image:     nil,
		}

		image, err := c.FormFile("image")
		if err == nil {
			file, err := image.Open()
			if err != nil {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Failed to open image file")
			}

			defer file.Close()

			req.Image = &requests.File{
				Name:        image.Filename,
				Size:        image.Size,
				Reader:      file,
				ContentType: image.Header.Get("Content-Type"),
			}
		}

		updatedSection, err := sectionService.UpdateByID(c.Context(), id, req)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"id":         updatedSection.ID,
			"name":       updatedSection.Name,
			"image":      updatedSection.Image,
			"started_at": updatedSection.StartedAt,
			"ended_at":   updatedSection.EndedAt,
		})
	})

	cmsSectionRouter.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		section, err := sectionService.GetByID(c.Context(), id)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"id":         section.ID,
			"name":       section.Name,
			"image":      section.Image,
			"started_at": section.StartedAt,
			"ended_at":   section.EndedAt,
		})
	})

	cmsSectionRouter.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := sectionService.DeleteByID(c.Context(), id); err != nil {
			return err
		}

		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{})
	})
}
