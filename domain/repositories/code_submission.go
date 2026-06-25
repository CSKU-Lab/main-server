package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type CodeSubmissionRepository interface {
	Create(ctx context.Context, payload *CreateCodeSubmissionPayload) error
	Update(ctx context.Context, payload *UpdateCodeSubmissionPayload) error
	Get(ctx context.Context, submissionID string) (*models.CodeSubmission, error)
	GetByIDs(ctx context.Context, submissionIDs []string) ([]*models.CodeSubmission, error)
}

type CreateCodeSubmissionPayload struct {
	SubmissionID string
	Files        models.SubmissionFiles
	StudentFiles models.SubmissionFiles
	RunnerID     string
}

type UpdateCodeSubmissionPayload struct {
	SubmissionID   string
	Status         string
	AvgWallTime    float32
	AvgMemory      int32
	TestCaseGroups models.TestCaseGroupResults
}
