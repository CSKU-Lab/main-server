package repositories

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type UpdateSubmissionRequest struct {
	ID          string
	Status      *models.SubmissionStatus
	AutoScore   *int
	ManualScore *int
}

type SubmissionRepository interface {
	Create(ctx context.Context, req *Submission) error
	Update(ctx context.Context, req *UpdateSubmissionRequest) error
	Get(ctx context.Context, id string) (*Submission, error)
	GetPagination(ctx context.Context, userID string, materialID string, page int, pageSize int, sortOrder string) ([]Submission, error)
	GetLatestByMaterialID(ctx context.Context, materialID string) ([]models.RawSubmission, error)
	GetLatestByMaterialAndStudentID(ctx context.Context, materialID string, studentID string) (*models.RawSubmission, error)

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
	CreatedAt  time.Time
}
