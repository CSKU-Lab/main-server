package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type CreateTypingSubmissionPayload struct {
	SubmissionID string
	RawWPM       float64
	AdjustedWPM  float64
	ErrorRate    float64
	Duration     float64
}

type BestTypingSubmission struct {
	SubmissionID string
	*models.TypingSubmission
}

type TypingSubmissionRepository interface {
	Create(ctx context.Context, payload *CreateTypingSubmissionPayload) error
	Get(ctx context.Context, submissionID string) (*models.TypingSubmission, error)
	GetByIDs(ctx context.Context, submissionIDs []string) (map[string]*models.TypingSubmission, error)
	GetBestByUserID(ctx context.Context, userID, materialID, labID, sectionID string) (*BestTypingSubmission, error)
	GetBestByMaterial(ctx context.Context, materialID, labID, sectionID string) ([]map[string]interface{}, error)
}
