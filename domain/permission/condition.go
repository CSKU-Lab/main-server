package permission

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

// adminCondition checks if a user has admin privileges.
type adminCondition struct{}

// Evaluate checks whether the user has the admin role.
func (c *adminCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	if user == nil {
		return false
	}
	for _, role := range user.Roles {
		if role == models.ADMIN {
			return true
		}
	}
	return false
}

// courseCreatorCondition checks if a user is a creator of a specific course.
type courseCreatorCondition struct {
	repo          repositories.CourseCreatorRepository
	courseIDParam string
}

// Evaluate checks whether the user is a creator of the course specified in params.
func (c *courseCreatorCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	if user == nil {
		return false
	}

	courseID, ok := params[c.courseIDParam]
	if !ok || courseID == "" {
		return false
	}

	creators, err := c.repo.GetCreators(ctx, courseID)
	if err != nil {
		return false
	}

	for _, creator := range creators {
		if creator.ID == user.ID {
			return true
		}
	}
	return false
}

// sectionInstructorCondition checks if a user is an instructor of a specific section.
type sectionInstructorCondition struct {
	repo           repositories.SectionInstructorRepository
	sectionIDParam string
}

// Evaluate checks whether the user is an instructor of the section specified in params.
func (c *sectionInstructorCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	if user == nil {
		return false
	}

	sectionID, ok := params[c.sectionIDParam]
	if !ok || sectionID == "" {
		return false
	}

	instructors, err := c.repo.Get(ctx, sectionID)
	if err != nil {
		return false
	}

	for _, instructor := range instructors {
		if instructor.ID == user.ID {
			return true
		}
	}
	return false
}

// sectionStudentCondition checks if a user is a student in a specific section.
type sectionStudentCondition struct {
	repo           repositories.SectionStudentRepository
	sectionIDParam string
}

// Evaluate checks whether the user is a student in the section specified in params.
func (c *sectionStudentCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	if user == nil {
		return false
	}

	sectionID, ok := params[c.sectionIDParam]
	if !ok || sectionID == "" {
		return false
	}

	// Use GetBySectionAndStudentID for efficient lookup
	_, err := c.repo.GetBySectionAndStudentID(ctx, sectionID, user.ID)
	return err == nil
}

// submissionOwnerCondition checks if a user owns a specific submission.
type submissionOwnerCondition struct {
	repo              repositories.SubmissionRepository
	submissionIDParam string
}

// Evaluate checks whether the user owns the submission specified in params.
func (c *submissionOwnerCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	if user == nil {
		return false
	}

	submissionID, ok := params[c.submissionIDParam]
	if !ok || submissionID == "" {
		return false
	}

	submission, err := c.repo.Get(ctx, submissionID)
	if err != nil {
		return false
	}

	return submission.UserID == user.ID
}

// authenticatedCondition checks if a user is authenticated (always true if user exists).
type authenticatedCondition struct{}

// Evaluate checks whether the user is authenticated (non-nil).
func (c *authenticatedCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	return user != nil
}

// andCondition implements AND logic as a Condition.
// It passes if ALL of its sub-conditions are satisfied.
type andCondition struct {
	conditions []Condition
}

// Evaluate returns true if ALL of the sub-conditions are satisfied.
func (a *andCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	for _, c := range a.conditions {
		if !c.Evaluate(ctx, user, params) {
			return false
		}
	}
	return true
}

// notCondition implements NOT logic as a Condition.
// It passes if its sub-condition is NOT satisfied.
type notCondition struct {
	condition Condition
}

// Evaluate returns true if the sub-condition is NOT satisfied.
func (n *notCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	return !n.condition.Evaluate(ctx, user, params)
}
