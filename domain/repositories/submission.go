package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type Submission interface {
	Create(ctx context.Context, req *SubmissionPayload) error
	Update(ctx context.Context, req *SubmissionPayload) error
	Get(ctx context.Context, ID string) (*models.Submission, error)
}

type SubmissionPayload struct {
	ID         string
	UserID     string
	MaterialID string
	SectionID  *string
	CourseID   *string
}
