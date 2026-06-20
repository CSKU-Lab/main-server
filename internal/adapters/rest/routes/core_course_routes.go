package routes

import (
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/gofiber/fiber/v3"
)

type myCourseInstructor struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
}

type myCourseResponse struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Description   *string              `json:"description"`
	Banner        *string              `json:"banner"`
	Visibility    string               `json:"visibility"`
	TotalStudents int                  `json:"total_students"`
	Instructors   []myCourseInstructor `json:"instructors"`
}

func NewCoreCourseRoute(
	router fiber.Router,
	courseService services.CourseService,
	enrollmentService services.CourseEnrollmentService,
	defaultLabService services.DefaultLabService,
	labMaterialService services.LabMaterialService,
	sectionService services.SectionService,
) {
	courseRouter := router.Group("/courses")

	courseRouter.Get("/featured", func(c fiber.Ctx) error {
		limitQuery := c.Query("limit", "4")
		limit, err := strconv.Atoi(limitQuery)
		if err != nil || limit < 1 {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid limit"})
		}

		courses, err := courseService.GetFeatured(c.RequestCtx(), limit)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{"data": courses})
	})

	courseRouter.Get("/", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "asc")
		show := c.Query("show", "active")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		courses, err := courseService.GetPagination(c.RequestCtx(), page, pageSize, search, sortBy, sortOrder, show, "public")
		if err != nil {
			return err
		}

		count, err := courseService.Count(c.RequestCtx(), search, show, "public")
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting courses count"})
		}

		totalPage := 1
		if pageSize > 0 {
			totalPage = int(math.Ceil(float64(count) / float64(pageSize)))
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": totalPage,
				"total_rows": count,
			},
			"data": courses,
		})
	})

	router.Get("/my-courses", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "asc")
		show := c.Query("show", "active")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		courses, err := courseService.GetPaginationForStudent(c.RequestCtx(), user.ID, page, pageSize, search, sortBy, sortOrder, show)
		if err != nil {
			return err
		}

		count, err := courseService.CountForStudent(c.RequestCtx(), user.ID, search, show)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting courses count"})
		}

		data := make([]myCourseResponse, 0, len(courses))
		for _, course := range courses {
			resp := myCourseResponse{
				ID:            course.ID,
				Name:          course.Name,
				Description:   course.Description,
				Banner:        course.Banner,
				Visibility:    course.Visibility,
				TotalStudents: course.TotalStudents,
				Instructors:   make([]myCourseInstructor, 0),
			}

			if course.Visibility == "public" {
				for _, cr := range course.Creators {
					resp.Instructors = append(resp.Instructors, myCourseInstructor{
						ID:           cr.ID,
						Username:     cr.Username,
						DisplayName:  cr.DisplayName,
						ProfileImage: cr.ProfileImage,
					})
				}
			} else {
				section, err := sectionService.GetByID(c.RequestCtx(), course.ID)
				if err != nil {
					return err
				}
				resp.Banner = section.Banner
				for _, inst := range section.Instructors {
					resp.Instructors = append(resp.Instructors, myCourseInstructor{
						ID:           inst.ID,
						Username:     inst.Username,
						DisplayName:  inst.DisplayName,
						ProfileImage: inst.ProfileImage,
					})
				}
			}

			data = append(data, resp)
		}

		totalPage := 1
		if pageSize > 0 {
			totalPage = int(math.Ceil(float64(count) / float64(pageSize)))
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": totalPage,
				"total_rows": count,
			},
			"data": data,
		})
	})

	courseRouter.Get("/:courseID", func(c fiber.Ctx) error {
		courseID := c.Params("courseID")

		course, err := courseService.GetByID(c.RequestCtx(), courseID)
		if err != nil {
			return err
		}

		if course.Visibility != "public" {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusForbidden, Message: "Course is not public"})
		}

		filterParams := map[string]string{
			"course_id__is": courseID,
		}
		labs, err := defaultLabService.GetPagination(c.RequestCtx(), 1, -1, "", "position", "asc", filterParams)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"course": course,
			"labs":   labs,
		})
	})

	courseRouter.Post("/:courseID/enroll", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")

		if err := enrollmentService.Enroll(c.RequestCtx(), courseID, user.ID); err != nil {
			return err
		}

		return c.SendStatus(http.StatusNoContent)
	})

	courseRouter.Delete("/:courseID/enroll", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")

		if err := enrollmentService.Unenroll(c.RequestCtx(), courseID, user.ID); err != nil {
			return err
		}

		return c.SendStatus(http.StatusNoContent)
	})

	courseRouter.Get("/:courseID/labs", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")

		enrolled, err := enrollmentService.IsEnrolled(c.RequestCtx(), courseID, user.ID)
		if err != nil {
			return err
		}
		if !enrolled {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusForbidden, Message: "Not enrolled in this course"})
		}

		filterParams := map[string]string{
			"course_id__is": courseID,
		}

		labs, err := defaultLabService.GetPagination(c.RequestCtx(), 1, -1, "", "position", "asc", filterParams)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{"data": labs})
	})

	courseRouter.Get("/:courseID/labs/:labID/materials", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")
		labID := c.Params("labID")

		enrolled, err := enrollmentService.IsEnrolled(c.RequestCtx(), courseID, user.ID)
		if err != nil {
			return err
		}
		if !enrolled {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusForbidden, Message: "Not enrolled in this course"})
		}

		materials, err := labMaterialService.GetByLabID(c.RequestCtx(), labID)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{"data": materials})
	})
}
