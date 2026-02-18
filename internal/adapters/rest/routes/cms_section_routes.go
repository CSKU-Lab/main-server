package routes

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCmsSectionRoutes(router fiber.Router, sectionService services.SectionService, semesterService services.SemesterService, labSectionService services.LabSectionService, sectionLogService services.SectionLogService, labService services.LabService, submissionService services.SubmissionService) {
	cmsSectionRouter := router.Group("/sections", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	cmsSectionRouter.Post("/", func(c fiber.Ctx) error {
		req := &requests.CreateSection{
			Name:       c.FormValue("name"),
			SemesterID: c.FormValue("semester_id"),
			CourseID:   c.FormValue("course_id"),
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

		form, err := c.MultipartForm()
		if err != nil {
			return err
		}

		instructors := form.Value["instructors[]"]
		instructorList := []string{}
		for _, instructor := range instructors {
			instructorList = append(instructorList, strings.TrimSpace(instructor))
		}

		if len(instructorList) > 0 {
			req.Instructors = instructorList
		}

		studentUsernames := form.Value["students[]"]
		studentList := []string{}
		for _, student := range studentUsernames {
			studentList = append(studentList, strings.TrimSpace(student))
		}

		if len(studentList) > 0 {
			req.Students = studentList
		}

		if err := req.Validate(); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":  "Bad Request",
				"fields": err,
			})
		}

		ctx := c.Context()

		ID, err := sectionService.Create(ctx, req)
		if err != nil {
			return err
		}

		err = sectionService.SetDefaultLabs(ctx, ID, req.CourseID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": ID,
		})
	})

	cmsSectionRouter.Patch("/:id", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		id := c.Params("id")
		req := &requests.UpdateSection{
			Name:       c.FormValue("name"),
			SemesterID: c.FormValue("semester_id"),
		}

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

		form, err := c.MultipartForm()
		if err != nil {
			return nil
		}

		instructors := form.Value["instructors[]"]
		req.Instructors = instructors

		students := form.Value["students[]"]
		req.Students = students

		ctx := c.Context()
		err = sectionService.UpdateByID(ctx, id, req, user.ID)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusAccepted)
	})

	// need to be refactored because this violates clean architecture
	cmsSectionRouter.Get("/", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
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

		filterParams := make(map[string]string)
		for key, value := range c.Queries() {
			if strings.Contains(key, "__") {
				filterParams[key] = value
			}
		}

		sems, err := semesterService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: err.Error()})
		}

		count, err := semesterService.Count(c.RequestCtx(), search, nil)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semesters count"})
		}

		type semesterFields struct {
			Name string              `json:"name"`
			Type models.SemesterType `json:"type"`
		}

		type sectionsOfSemester struct {
			Semester semesterFields   `json:"semester"`
			Sections []models.Section `json:"sections"`
		}

		sectionsOfSemesters := make([]sectionsOfSemester, len(sems))
		for i, semester := range sems {
			sections, err := sectionService.GetBySemesterID(c.RequestCtx(), semester.ID)
			if err != nil {
				return err
			}

			responseSections := []models.Section{}
			if len(sections) > 0 {
				responseSections = sections
			}

			sectionsOfSemesters[i] = sectionsOfSemester{
				Semester: semesterFields{
					Name: semester.Name,
					Type: semester.Type,
				},
				Sections: responseSections,
			}
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
				"total_rows": count,
			},
			"data": sectionsOfSemesters,
		})
	})

	cmsSectionRouter.Get("/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		section, err := sectionService.GetByID(c.RequestCtx(), id)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(section)
	})

	cmsSectionRouter.Delete("/:id", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		id := c.Params("id")

		ctx := c.Context()
		if err := sectionService.DeleteByID(ctx, id, user.ID); err != nil {
			return err
		}

		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{})
	})

	cmsSectionRouter.Get(":id/students", func(c fiber.Ctx) error {
		sectionID := c.Params("id")

		students, err := sectionService.GetStudentsBySectionID(c.RequestCtx(), sectionID)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": students,
		})
	})

	cmsSectionRouter.Post(":id/students", middlewares.ValidateMiddleware[requests.SectionStudents](), func(c fiber.Ctx) error {
		sectionID := c.Params("id")
		body := c.Locals("body").(*requests.SectionStudents)

		ctx := c.Context()
		err := sectionService.AddStudents(ctx, sectionID, body.StudentUsernames)
		if err != nil {
			log.Println(err)
			return err
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Students added successfully",
		})
	})

	cmsSectionRouter.Post(":id/students/remove", middlewares.ValidateMiddleware[requests.RemoveStudent](), func(c fiber.Ctx) error {
		sectionID := c.Params("id")
		body := c.Locals("body").(*requests.RemoveStudent)

		ctx := c.Context()
		err := sectionService.RemoveStudents(ctx, sectionID, body.StudentIDs)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Students removed successfully",
		})
	})

	cmsSectionRouter.Get("/:sectionID/labs", func(c fiber.Ctx) error {
		sectionID := c.Params("sectionID")

		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		sortBy := c.Query("sort_by", "position")
		sortOrder := c.Query("sort_order", "asc")

		filterParams := make(map[string]string)
		for key, value := range c.Queries() {
			if strings.Contains(key, "__") {
				filterParams[key] = value
			}
		}

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		filterParams["section_id__is"] = sectionID

		labSections, err := labSectionService.GetPagination(c.RequestCtx(), page, pageSize, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := labSectionService.Count(c.RequestCtx(), filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting labs count"})
		}

		type labSectionResponse struct {
			LabID    string     `json:"lab_id"`
			Position int        `json:"position"`
			Status   string     `json:"status"`
			OpenedAt *time.Time `json:"opened_at"`
			ClosedAt *time.Time `json:"closed_at"`
			LabName  string     `json:"lab_name"`
		}

		responseSections := make([]labSectionResponse, len(labSections))
		for i, section := range labSections {
			lab, err := labService.GetByID(c.RequestCtx(), section.LabID)
			if err != nil {
				return err
			}

			responseSections[i] = labSectionResponse{
				LabID:    section.LabID,
				Position: section.Position,
				Status:   section.Status,
				OpenedAt: section.OpenedAt,
				ClosedAt: section.ClosedAt,
				LabName:  lab.DisplayName,
			}
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": responseSections,
		})
	})

	cmsSectionRouter.Get("/:sectionID/labs/:labID", func(c fiber.Ctx) error {
		sectionID := c.Params("sectionID")
		labID := c.Params("labID")

		// Get lab section details
		labSection, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), labID, sectionID)
		if err != nil {
			return err
		}

		// Get lab name
		lab, err := labService.GetByID(c.RequestCtx(), labID)
		if err != nil {
			return err
		}

		// Get total students
		students, err := sectionService.GetStudentsBySectionID(c.RequestCtx(), sectionID)
		if err != nil {
			return err
		}
		totalStudents := len(students)

		// Get completed students (passed all materials)
		completedStudents, err := submissionService.CountCompletedStudentsByLabAndSection(c.RequestCtx(), labID, sectionID)
		if err != nil {
			return err
		}

		type labDetailResponse struct {
			LabName           string     `json:"lab_name"`
			Status            string     `json:"status"`
			OpenedAt          *time.Time `json:"opened_at"`
			ClosedAt          *time.Time `json:"closed_at"`
			CompletedStudents int        `json:"completed_students"`
			TotalStudents     int        `json:"total_students"`
		}

		return c.JSON(labDetailResponse{
			LabName:           lab.DisplayName,
			Status:            labSection.Status,
			OpenedAt:          labSection.OpenedAt,
			ClosedAt:          labSection.ClosedAt,
			CompletedStudents: completedStudents,
			TotalStudents:     totalStudents,
		})
	})

	cmsSectionRouter.Post("/:sectionID/labs", middlewares.ValidateMiddleware[requests.SetLabSection](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		sectionID := c.Params("sectionID")
		req := c.Locals("body").(*requests.SetLabSection)

		ctx := c.Context()
		err := labSectionService.Create(ctx, req, user.ID, sectionID)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	cmsSectionRouter.Patch("/:sectionID/labs", middlewares.ValidateMiddleware[requests.UpdateLabSection](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		sectionID := c.Params("sectionID")
		req := c.Locals("body").(*requests.UpdateLabSection)

		ctx := c.Context()
		return labSectionService.Update(ctx, user.ID, sectionID, req)
	})

	cmsSectionRouter.Post("/:sectionID/labs/delete", middlewares.ValidateMiddleware[requests.DeleteLabSection](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		sectionID := c.Params("sectionID")
		req := c.Locals("body").(*requests.DeleteLabSection)

		ctx := c.Context()
		return labSectionService.Delete(ctx, sectionID, user.ID, req)
	})

	cmsSectionRouter.Get("/:sectionID/logs", func(c fiber.Ctx) error {
		sectionID := c.Params("sectionID")

		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "timestamp")
		sortOrder := c.Query("sort_order", "desc")

		filterParams := make(map[string]string)
		for key, value := range c.Queries() {
			if strings.Contains(key, "__") {
				filterParams[key] = value
			}
		}

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		logs, err := sectionLogService.GetPaginationBySectionID(c.RequestCtx(), sectionID, page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := sectionLogService.CountBySectionID(c.RequestCtx(), sectionID, search, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semesters count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count/pageSize) + 1),
				"total_rows": count,
			},
			"data": logs,
		})
	})

	cmsSectionRouter.Get("/:id/gradebook", func(c fiber.Ctx) error {
		id := c.Params("id")
		gradebook, err := submissionService.GetGradebookBySectionID(c.RequestCtx(), id)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusOK).JSON(gradebook)
	})
}

// fiber:context-methods migrated
