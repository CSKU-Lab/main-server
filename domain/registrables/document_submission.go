package registrables

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

// DocumentSubmission is the SubmissionRegistrable for document-type materials.
// Document submissions do not have a per-submission payload table; all relevant
// data lives in the base submissions row, so every payload-returning method
// returns nil / empty maps without error.
type DocumentSubmission struct{}

func NewDocumentSubmission() *DocumentSubmission {
	return &DocumentSubmission{}
}

func (d *DocumentSubmission) Create(_ context.Context, _ repositories.UoWInstance, _ string, _ string, _ []byte) error {
	return nil
}

func (d *DocumentSubmission) Update(_ context.Context, _ repositories.UoWInstance, _ string, _ []byte) error {
	return nil
}

func (d *DocumentSubmission) Get(_ context.Context, _ string, _ string) (any, error) {
	return nil, nil
}

func (d *DocumentSubmission) GetByIDs(_ context.Context, submissionIDs []string, _ string) (map[string]any, error) {
	result := make(map[string]any, len(submissionIDs))
	for _, id := range submissionIDs {
		result[id] = nil
	}
	return result, nil
}

func (d *DocumentSubmission) GetOverviewsPayload(_ context.Context, submissionIDs []string) (map[string]any, error) {
	result := make(map[string]any, len(submissionIDs))
	for _, id := range submissionIDs {
		result[id] = nil
	}
	return result, nil
}

func (d *DocumentSubmission) GetOverviewStats(payload any) any {
	return nil
}

func (d *DocumentSubmission) GetOverviewStatsByID(_ context.Context, _ string) any {
	return nil
}
