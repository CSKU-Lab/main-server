package middlewares

import (
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/gofiber/fiber/v3"
)

// PermissionBuilder provides a fluent API for building permission middleware.
// It allows chaining methods to specify resource types, parameters, and actions.
type PermissionBuilder struct {
	permService  permission.Service
	resourceType string
	paramKey     string
	action       string
	fromQuery    bool
	fromLocals   bool
}

// FromQuery makes the builder read the resource ID from a query parameter instead of a URL param.
func (pb *PermissionBuilder) FromQuery() *PermissionBuilder {
	pb.fromQuery = true
	return pb
}

// FromLocals makes the builder read the resource ID from c.Locals instead of a URL param.
func (pb *PermissionBuilder) FromLocals() *PermissionBuilder {
	pb.fromLocals = true
	return pb
}

// getResourceID reads the resource ID from the appropriate source.
func (pb *PermissionBuilder) getResourceID(c fiber.Ctx) string {
	if pb.fromLocals {
		if id, ok := c.Locals(pb.paramKey).(string); ok {
			return id
		}
		return ""
	}
	if pb.fromQuery {
		return c.Query(pb.paramKey)
	}
	return c.Params(pb.paramKey)
}

// Permission creates a new PermissionBuilder with the given permission service.
// This is the entry point for the fluent builder pattern.
func Permission(permService permission.Service) *PermissionBuilder {
	return &PermissionBuilder{
		permService: permService,
	}
}

// ForSection sets the resource type to "section" and specifies the URL parameter key.
func (pb *PermissionBuilder) ForSection(paramKey string) *PermissionBuilder {
	pb.resourceType = "section"
	pb.paramKey = paramKey
	return pb
}

// ForCourse sets the resource type to "course" and specifies the URL parameter key.
func (pb *PermissionBuilder) ForCourse(paramKey string) *PermissionBuilder {
	pb.resourceType = "course"
	pb.paramKey = paramKey
	return pb
}

// ForLab sets the resource type to "lab" and specifies the URL parameter key.
func (pb *PermissionBuilder) ForLab(paramKey string) *PermissionBuilder {
	pb.resourceType = "lab"
	pb.paramKey = paramKey
	return pb
}

// ForMaterial sets the resource type to "material" and specifies the URL parameter key.
func (pb *PermissionBuilder) ForMaterial(paramKey string) *PermissionBuilder {
	pb.resourceType = "material"
	pb.paramKey = paramKey
	return pb
}

// ForSubmission sets the resource type to "submission" and specifies the URL parameter key.
func (pb *PermissionBuilder) ForSubmission(paramKey string) *PermissionBuilder {
	pb.resourceType = "submission"
	pb.paramKey = paramKey
	return pb
}

// CanCreate returns a middleware that checks if the user can create the resource.
// Admin, course creators, and section instructors can create resources.
func (pb *PermissionBuilder) CanCreate() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		resourceID := pb.getResourceID(c)

		// Admin can always create
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Check resource-specific permissions
		switch pb.resourceType {
		case "section":
			// Section creation requires course creator or instructor status
			if resourceID != "" {
				isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, resourceID)
				if err != nil {
					return err
				}
				if isCreator {
					return c.Next()
				}
			}
		case "course":
			// Only admin can create courses (already checked above)
		case "lab":
			// Lab creation requires course creator status
			if resourceID != "" {
				isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, resourceID)
				if err != nil {
					return err
				}
				if isCreator {
					return c.Next()
				}
			}
		case "material":
			// Material creation requires section instructor status
			if resourceID != "" {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, resourceID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot create %s. Only admin or authorized users can create %s.", pb.resourceType, pb.resourceType),
		})
	}
}

// CanView returns a middleware that checks if the user can view the resource.
// Admin, course creators, section instructors, and enrolled students can view.
func (pb *PermissionBuilder) CanView() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		resourceID := pb.getResourceID(c)

		// Admin can always view
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Check resource-specific permissions
		switch pb.resourceType {
		case "section":
			// Section viewing requires being an instructor or student
			isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, resourceID)
			if err != nil {
				return err
			}
			if isInstructor {
				return c.Next()
			}

			isStudent, err := pb.permService.IsSectionStudent(c.Context(), user.ID, resourceID)
			if err != nil {
				return err
			}
			if isStudent {
				return c.Next()
			}
		case "course":
			// Course viewing requires being a creator
			isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, resourceID)
			if err != nil {
				return err
			}
			if isCreator {
				return c.Next()
			}
		case "lab":
			// Lab viewing requires section membership
			sectionID := c.Params("section_id")
			if sectionID != "" {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, sectionID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}

				isStudent, err := pb.permService.IsSectionStudent(c.Context(), user.ID, sectionID)
				if err != nil {
					return err
				}
				if isStudent {
					return c.Next()
				}
			}
		case "material":
			// Material viewing requires section membership
			sectionID := c.Params("section_id")
			if sectionID != "" {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, sectionID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}

				isStudent, err := pb.permService.IsSectionStudent(c.Context(), user.ID, sectionID)
				if err != nil {
					return err
				}
				if isStudent {
					return c.Next()
				}
			}
		case "submission":
			// Submission viewing requires being the owner or an instructor
			isOwner, err := pb.permService.IsSubmissionOwner(c.Context(), user.ID, resourceID)
			if err != nil {
				return err
			}
			if isOwner {
				return c.Next()
			}

			// Check if user is instructor of the section where submission was made
			submission, err := pb.permService.GetSubmission(c.Context(), resourceID)
			if err != nil {
				return err
			}
			if submission.SectionID != nil {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, *submission.SectionID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot view %s %s.", pb.resourceType, resourceID),
		})
	}
}

// CanModify returns a middleware that checks if the user can modify the resource.
// Admin, course creators (for their courses), and section instructors can modify.
func (pb *PermissionBuilder) CanModify() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		resourceID := pb.getResourceID(c)

		// Admin can always modify
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Check resource-specific permissions
		switch pb.resourceType {
		case "section":
			// Section modification requires being an instructor
			isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, resourceID)
			if err != nil {
				return err
			}
			if isInstructor {
				return c.Next()
			}

			// Or being a course creator
			section, err := pb.permService.GetSection(c.Context(), resourceID)
			if err != nil {
				return err
			}
			isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, section.CourseID)
			if err != nil {
				return err
			}
			if isCreator {
				return c.Next()
			}
		case "course":
			// Course modification requires being a creator
			isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, resourceID)
			if err != nil {
				return err
			}
			if isCreator {
				return c.Next()
			}
		case "lab":
			// Lab modification requires course creator status
			courseID := c.Params("course_id")
			if courseID != "" {
				isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, courseID)
				if err != nil {
					return err
				}
				if isCreator {
					return c.Next()
				}
			}
		case "material":
			// Material modification requires section instructor status
			sectionID := c.Params("section_id")
			if sectionID != "" {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, sectionID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}
			}
		case "submission":
			// Submission modification (e.g., updating scores) requires instructor status
			submission, err := pb.permService.GetSubmission(c.Context(), resourceID)
			if err != nil {
				return err
			}
			if submission.SectionID != nil {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, *submission.SectionID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot modify %s %s. Only admin or authorized users can modify.", pb.resourceType, resourceID),
		})
	}
}

// CanDelete returns a middleware that checks if the user can delete the resource.
// Admin and course creators can delete sections; only admin can delete courses.
func (pb *PermissionBuilder) CanDelete() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		resourceID := pb.getResourceID(c)

		// Admin can always delete
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Check resource-specific permissions
		switch pb.resourceType {
		case "section":
			// Section deletion requires being a course creator
			section, err := pb.permService.GetSection(c.Context(), resourceID)
			if err != nil {
				return err
			}
			isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, section.CourseID)
			if err != nil {
				return err
			}
			if isCreator {
				return c.Next()
			}
		case "course":
			// Only admin can delete courses (already checked above)
		case "lab":
			// Lab deletion requires course creator status
			courseID := c.Params("course_id")
			if courseID != "" {
				isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, courseID)
				if err != nil {
					return err
				}
				if isCreator {
					return c.Next()
				}
			}
		case "material":
			// Material deletion requires section instructor status
			sectionID := c.Params("section_id")
			if sectionID != "" {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, sectionID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot delete %s %s. Only admin or course creator can delete %s.", pb.resourceType, resourceID, pb.resourceType),
		})
	}
}

// CanGrade returns a middleware that checks if the user can grade submissions.
// Admin and section instructors can grade.
func (pb *PermissionBuilder) CanGrade() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		resourceID := pb.getResourceID(c)

		// Admin can always grade
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Check resource-specific permissions
		switch pb.resourceType {
		case "section":
			// Grading in a section requires instructor status
			isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, resourceID)
			if err != nil {
				return err
			}
			if isInstructor {
				return c.Next()
			}
		case "submission":
			// Grading a submission requires being instructor of the section
			submission, err := pb.permService.GetSubmission(c.Context(), resourceID)
			if err != nil {
				return err
			}
			if submission.SectionID != nil {
				isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, *submission.SectionID)
				if err != nil {
					return err
				}
				if isInstructor {
					return c.Next()
				}
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot grade %s %s. Only admin or section instructor can grade.", pb.resourceType, resourceID),
		})
	}
}

// CanCreateLab returns a middleware that checks if the user can create labs.
// Admin and course creators can create labs.
func (pb *PermissionBuilder) CanCreateLab() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params(pb.paramKey)

		// Admin can always create labs
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Course creators can create labs for their courses
		if courseID != "" {
			isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, courseID)
			if err != nil {
				return err
			}
			if isCreator {
				return c.Next()
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot create lab in course %s. Only admin or course creator can create labs.", courseID),
		})
	}
}

// CanManageStudents returns a middleware that checks if the user can manage students in a section.
// Admin and section instructors can manage students.
func (pb *PermissionBuilder) CanManageStudents() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		sectionID := c.Params(pb.paramKey)

		// Admin can always manage students
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Section instructors can manage students
		if sectionID != "" {
			isInstructor, err := pb.permService.IsSectionInstructor(c.Context(), user.ID, sectionID)
			if err != nil {
				return err
			}
			if isInstructor {
				return c.Next()
			}

			// Course creators can also manage students
			section, err := pb.permService.GetSection(c.Context(), sectionID)
			if err != nil {
				return err
			}
			isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, section.CourseID)
			if err != nil {
				return err
			}
			if isCreator {
				return c.Next()
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot manage students in section %s. Only admin, course creator, or section instructor can manage students.", sectionID),
		})
	}
}

// CanManageInstructors returns a middleware that checks if the user can manage instructors in a section.
// Admin and course creators can manage instructors.
func (pb *PermissionBuilder) CanManageInstructors() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		sectionID := c.Params(pb.paramKey)

		// Admin can always manage instructors
		isAdmin, err := pb.permService.IsAdmin(c.Context(), user.ID)
		if err != nil {
			return err
		}
		if isAdmin {
			return c.Next()
		}

		// Course creators can manage instructors for sections in their courses
		if sectionID != "" {
			section, err := pb.permService.GetSection(c.Context(), sectionID)
			if err != nil {
				return err
			}
			isCreator, err := pb.permService.IsCourseCreator(c.Context(), user.ID, section.CourseID)
			if err != nil {
				return err
			}
			if isCreator {
				return c.Next()
			}
		}

		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    fmt.Sprintf("Permission denied: cannot manage instructors in section %s. Only admin or course creator can manage instructors.", sectionID),
		})
	}
}
