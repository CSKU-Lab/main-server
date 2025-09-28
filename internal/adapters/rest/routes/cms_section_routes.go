package routes

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
)

func NewCmsSectionRoutes(router fiber.Router, sectionService services.SectionService, semesterService services.SemesterService) {
	cmsSectionRouter := router.Group("/sections")

	cmsSectionRouter.Post("/", func(c *fiber.Ctx) error {
		name := c.FormValue("name")
		courseID := c.FormValue("course_id")
		semesterID := c.FormValue("semester_id")
		instructors := c.FormValue("instructors")

		instructorList := []string{}
		if instructors != "" {
			for _, instructor := range strings.Split(instructors, ",") {
				instructorList = append(instructorList, strings.TrimSpace(instructor))
			}
		}

		req := &requests.CreateSection{
			Name:        name,
			Instructors: instructorList,
			SemesterID:  semesterID,
			CourseID:    courseID,
		}

		if err := req.Validate(); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":  "Bad Request",
				"fields": err,
			})
		}

		image, err := c.FormFile("banner")
		if err == nil {
			file, err := image.Open()
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to open image file"})
			}

			defer file.Close()

			imageFile := &requests.File{
				Name:        image.Filename,
				Size:        image.Size,
				Reader:      file,
				ContentType: image.Header.Get("Content-Type"),
			}

			req.Banner = imageFile
		}

		err = sectionService.Create(c.Context(), req)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	cmsSectionRouter.Patch("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		req := &requests.UpdateSection{
			Name:       c.FormValue("name"),
			SemesterID: c.FormValue("semester_id"),
		}

		instructors := c.FormValue("instructors")
		instructorList := []string{}
		if instructors != "" {
			for _, instructor := range strings.Split(instructors, ",") {
				instructorList = append(instructorList, strings.TrimSpace(instructor))
			}
		}

		req.Instructors = instructorList

		image, err := c.FormFile("banner")
		if err == nil {
			file, err := image.Open()
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to open image file"})
			}
			defer file.Close()

			req.Banner = &requests.File{
				Name:        image.Filename,
				Size:        image.Size,
				Reader:      file,
				ContentType: image.Header.Get("Content-Type"),
			}
		}

		err = sectionService.UpdateByID(c.Context(), id, req)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusAccepted)
	})

	cmsSectionRouter.Get("/", func(c *fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("pageSize", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "desc")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		sems, err := semesterService.GetPagination(c.Context(), page, pageSize, search, sortBy, sortOrder)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: err.Error()})
		}

		count, err := semesterService.Count(c.Context(), search)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semesters count"})
		}

		type sectionsOfSemester struct {
			SemesterName string           `json:"semester_name"`
			Sections     []models.Section `json:"sections"`
		}

		var sectionsOfSemesters []sectionsOfSemester
		for _, semester := range sems {
			sections, err := sectionService.GetBySemesterID(c.Context(), semester.ID)
			if err != nil {
				return err
			}

			sectionsOfSemesters = append(sectionsOfSemesters, sectionsOfSemester{
				SemesterName: semester.Name,
				Sections:     sections,
			})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": sectionsOfSemesters,
		})
	})

	cmsSectionRouter.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		section, err := sectionService.GetByID(c.Context(), id)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(section)
	})

	cmsSectionRouter.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := sectionService.DeleteByID(c.Context(), id); err != nil {
			return err
		}

		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{})
	})
}
