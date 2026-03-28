// Package permission provides a flexible, type-safe permission service for access control.
//
// The permission service uses a service pattern with composable conditions that can be
// used with middleware for route-level permission guards.
//
// Example usage:
//
//	permService := permission.NewPermissionService(adminRepo, courseCreatorRepo, ...)
//
//	// In route setup:
//	router.Get("/courses/:id",
//	    middleware.RequirePermission(permService.IsCourseCreator("id")),
//	    handler)
//
//	// Complex conditions:
//	router.Get("/admin",
//	    middleware.RequirePermission(
//	        permService.Or(
//	            permService.IsAdmin(),
//	            permService.IsCourseCreator("id"),
//	        ),
//	    ),
//	    handler)
package permission

import (
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

// Service provides permission checking capabilities with dependency injection.
// It uses existing repositories to verify permissions without creating new DB connections.
type Service interface {
	// Condition factories
	IsAdmin() Condition
	IsCourseCreator(courseIDParam string) Condition
	IsSectionInstructor(sectionIDParam string) Condition
	IsSectionStudent(sectionIDParam string) Condition
	IsSubmissionOwner(submissionIDParam string) Condition
	IsAuthenticated() Condition

	// Logical operators
	Or(conditions ...Condition) Condition
	And(conditions ...Condition) Condition
	Not(condition Condition) Condition
}

// permissionService implements the Service interface using repository pattern.
type permissionService struct {
	courseCreatorRepo     repositories.CourseCreatorRepository
	sectionInstructorRepo repositories.SectionInstructorRepository
	sectionStudentRepo    repositories.SectionStudentRepository
	submissionRepo        repositories.SubmissionRepository
}

// NewPermissionService creates a new permission service with the required repositories.
// All repositories are injected via constructor (DI pattern).
func NewPermissionService(
	courseCreatorRepo repositories.CourseCreatorRepository,
	sectionInstructorRepo repositories.SectionInstructorRepository,
	sectionStudentRepo repositories.SectionStudentRepository,
	submissionRepo repositories.SubmissionRepository,
) Service {
	return &permissionService{
		courseCreatorRepo:     courseCreatorRepo,
		sectionInstructorRepo: sectionInstructorRepo,
		sectionStudentRepo:    sectionStudentRepo,
		submissionRepo:        submissionRepo,
	}
}

// IsAdmin returns a condition that checks if the user has admin role.
func (s *permissionService) IsAdmin() Condition {
	return &adminCondition{}
}

// IsCourseCreator returns a condition that checks if the user is a creator of the specified course.
// The courseIDParam is the route parameter name (e.g., "id" for /courses/:id).
func (s *permissionService) IsCourseCreator(courseIDParam string) Condition {
	return &courseCreatorCondition{
		repo:          s.courseCreatorRepo,
		courseIDParam: courseIDParam,
	}
}

// IsSectionInstructor returns a condition that checks if the user is an instructor of the specified section.
// The sectionIDParam is the route parameter name (e.g., "id" for /sections/:id).
func (s *permissionService) IsSectionInstructor(sectionIDParam string) Condition {
	return &sectionInstructorCondition{
		repo:           s.sectionInstructorRepo,
		sectionIDParam: sectionIDParam,
	}
}

// IsSectionStudent returns a condition that checks if the user is a student in the specified section.
// The sectionIDParam is the route parameter name (e.g., "id" for /sections/:id).
func (s *permissionService) IsSectionStudent(sectionIDParam string) Condition {
	return &sectionStudentCondition{
		repo:           s.sectionStudentRepo,
		sectionIDParam: sectionIDParam,
	}
}

// IsSubmissionOwner returns a condition that checks if the user owns the specified submission.
// The submissionIDParam is the route parameter name (e.g., "id" for /submissions/:id).
func (s *permissionService) IsSubmissionOwner(submissionIDParam string) Condition {
	return &submissionOwnerCondition{
		repo:              s.submissionRepo,
		submissionIDParam: submissionIDParam,
	}
}

// IsAuthenticated returns a condition that always passes (user is authenticated).
// This is useful for routes that require login but no specific role.
func (s *permissionService) IsAuthenticated() Condition {
	return &authenticatedCondition{}
}

// Or combines multiple conditions with OR logic.
// The resulting condition passes if ANY of the input conditions are satisfied.
func (s *permissionService) Or(conditions ...Condition) Condition {
	return &orCondition{conditions: conditions}
}

// And combines multiple conditions with AND logic.
// The resulting condition passes if ALL of the input conditions are satisfied.
func (s *permissionService) And(conditions ...Condition) Condition {
	return &andCondition{conditions: conditions}
}

// Not negates a condition.
// The resulting condition passes if the input condition is NOT satisfied.
func (s *permissionService) Not(condition Condition) Condition {
	return &notCondition{condition: condition}
}
