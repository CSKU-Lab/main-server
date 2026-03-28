package sqlx

import (
	"context"
)

// PermissionRepositoryProvider implements the permission.RepositoryProvider interface.
// It provides database-backed permission checks.
type PermissionRepositoryProvider struct {
	db instance
}

// NewPermissionRepositoryProvider creates a new permission repository provider.
func NewPermissionRepositoryProvider(db instance) *PermissionRepositoryProvider {
	return &PermissionRepositoryProvider{db: db}
}

// IsSectionInstructor checks if a user is an instructor of a specific section.
func (p *PermissionRepositoryProvider) IsSectionInstructor(ctx context.Context, userID string, sectionID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM section_instructors 
			WHERE section_id = $1 AND instructor_id = $2
		)
	`
	var exists bool
	err := p.db.GetContext(ctx, &exists, query, sectionID, userID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// IsSubmissionOwner checks if a user owns a specific submission.
func (p *PermissionRepositoryProvider) IsSubmissionOwner(ctx context.Context, userID string, submissionID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM submissions 
			WHERE id = $1 AND user_id = $2
		)
	`
	var exists bool
	err := p.db.GetContext(ctx, &exists, query, submissionID, userID)
	if err != nil {
		return false, err
	}
	return exists, nil
}
