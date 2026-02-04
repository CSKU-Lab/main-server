package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type SubmissionRepository interface {
	Create(ctx context.Context, req *Submission) error
	Update(ctx context.Context, id string, status models.SubmissionStatus) error
	Get(ctx context.Context, id string) (*Submission, error)
}

type Submission struct {
	ID         string
	UserID     string
	LabID      string
	SectionID  *string
	CourseID   *string
	MaterialID string
	Status     models.SubmissionStatus
}
