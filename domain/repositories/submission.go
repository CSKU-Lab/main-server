package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type Submission interface {
	Create(ctx context.Context, req *SubmissionPayload) error
	Update(ctx context.Context, id string, status models.SubmissionStatus) error
	Get(ctx context.Context, id string) (*models.Submission, error)
}

type SubmissionPayload struct {
	ID         string
	UserID     string
	LabID      string
	SectionID  *string
	CourseID   *string
	MaterialID string
}
