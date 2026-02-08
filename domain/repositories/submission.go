package repositories

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type SubmissionRepository interface {
	Create(ctx context.Context, req *Submission) error
	Update(ctx context.Context, id string, status models.SubmissionStatus) error
	Get(ctx context.Context, id string) (*Submission, error)
	GetPagination(ctx context.Context, userID string, materialID string, page int, pageSize int, sortOrder string) ([]SubmissionOverview, error)
	Count(ctx context.Context, userID string, materialID string) (int, error)
}

type Submission struct {
	ID         string
	UserID     string
	LabID      string
	SectionID  *string
	CourseID   *string
	MaterialID string
	Status     models.SubmissionStatus
	Order      int
}

type SubmissionOverview struct {
	ID         string
	Status     models.SubmissionStatus
	MaterialID string
	CreatedAt  time.Time
}
