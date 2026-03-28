// Package permission provides the permission service interface for access control.
//
// This service provides methods to check user permissions across different resources
// including roles, courses, sections, and submissions. It is used by the middleware
// layer to enforce authorization rules.
package permission

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

// Service defines the interface for permission checks.
// Implementations must provide methods to verify user access rights across
// different resources in the system.
type Service interface {
	// Role checks
	IsAdmin(ctx context.Context, userID string) (bool, error)
	IsInstructor(ctx context.Context, userID string) (bool, error)
	IsStudent(ctx context.Context, userID string) (bool, error)

	// Course permissions
	IsCourseCreator(ctx context.Context, userID string, courseID string) (bool, error)
	IsCourseInstructor(ctx context.Context, userID string, courseID string) (bool, error)

	// Section permissions
	IsSectionInstructor(ctx context.Context, userID string, sectionID string) (bool, error)
	IsSectionStudent(ctx context.Context, userID string, sectionID string) (bool, error)
	GetSection(ctx context.Context, sectionID string) (*models.Section, error)

	// Submission permissions
	IsSubmissionOwner(ctx context.Context, userID string, submissionID string) (bool, error)
	GetSubmission(ctx context.Context, submissionID string) (*repositories.Submission, error)
}
