package registrables

import (
	"context"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type codeSubmission struct {
	repo repositories.CodeSubmission
}

func NewCodeSubmission(repo repositories.CodeSubmission) registries.SubmissionRegistrable {
	return &codeSubmission{
		repo: repo,
	}
}

type createCodeSubmissionPayload struct {
	Code string `json:"code"`
}

type updateCodeSubmissionPayload struct {
	Code           string                 `json:"code"`
	Status         string                 `json:"status"`
	AvgWallTime    float32                `json:"avg_wall_time"`
	AvgMemory      int32                  `json:"avg_memory"`
	TestCaseGroups []models.TestCaseGroup `json:"test_case_groups"`
}

func (c *codeSubmission) Create(ctx context.Context, uowRepo repositories.UoWInstance, submissionID string, payload []byte) error {
	parsedPayload, err := parsePayload[createCodeSubmissionPayload](payload)
	if err != nil {
		return errors.New("invalid payload type")
	}

	createPayload := &repositories.CreateCodeSubmissionPayload{
		SubmissionID: submissionID,
		Code:         parsedPayload.Code,
	}

	return uowRepo.CodeSubmission().Create(ctx, createPayload)
}

func (c *codeSubmission) Update(ctx context.Context, uowRepo repositories.UoWInstance, submissionID string, payload []byte) error {
	parsedPayload, err := parsePayload[updateCodeSubmissionPayload](payload)
	if err != nil {
		return errors.New("invalid payload type")
	}

	updatePayload := &repositories.UpdateCodeSubmissionPayload{
		SubmissionID:   submissionID,
		Status:         parsedPayload.Status,
		AvgWallTime:    parsedPayload.AvgWallTime,
		AvgMemory:      parsedPayload.AvgMemory,
		TestCaseGroups: parsedPayload.TestCaseGroups,
	}

	return uowRepo.CodeSubmission().Update(ctx, updatePayload)
}

func (c *codeSubmission) Get(ctx context.Context, submissionID string) (any, error) {
	return c.repo.Get(ctx, submissionID)
}
