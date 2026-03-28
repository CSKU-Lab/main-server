package middlewares

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/gofiber/fiber/v3"
)

// PermissionMiddleware creates a middleware that checks if the user satisfies the given permission check
func PermissionMiddleware(check func(*models.User, fiber.Ctx) error) func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		if err := check(user, c); err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    err.Error(),
			})
		}

		return c.Next()
	}
}

// IsAuthenticated checks if the user is authenticated (has a valid session)
// This is a basic check that passes for any logged-in user
func IsAuthenticated() func(*models.User, fiber.Ctx) error {
	return func(user *models.User, c fiber.Ctx) error {
		// User is already authenticated if they reach this point
		// (ProtectedRouteMiddleware runs before this)
		return nil
	}
}

// IsSectionStudent checks if the user is enrolled in the specified section
// paramName is the URL parameter name that contains the section ID (e.g., "sectionID", "id")
func IsSectionStudent(sectionStudentService services.SectionStudentService, paramName string) func(*models.User, fiber.Ctx) error {
	return func(user *models.User, c fiber.Ctx) error {
		sectionID := c.Params(paramName)
		if sectionID == "" {
			// Try to get from query parameter if not in path
			sectionID = c.Query(paramName, "")
		}
		if sectionID == "" {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Section ID is required",
			})
		}

		_, err := sectionStudentService.GetBySectionAndStudentID(c.RequestCtx(), sectionID, user.ID)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "You are not enrolled in this section",
			})
		}

		return nil
	}
}

// IsSubmissionOwner checks if the user owns the specified submission
// paramName is the URL parameter name that contains the submission ID (e.g., "submissionId", "id")
func IsSubmissionOwner(submissionRepo repositories.SubmissionRepository, paramName string) func(*models.User, fiber.Ctx) error {
	return func(user *models.User, c fiber.Ctx) error {
		submissionID := c.Params(paramName)
		if submissionID == "" {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Submission ID is required",
			})
		}

		submission, err := submissionRepo.Get(c.RequestCtx(), submissionID)
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusNotFound,
				Message:    "Submission not found",
			})
		}

		if submission.UserID != user.ID {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "You can only access your own submissions",
			})
		}

		return nil
	}
}
