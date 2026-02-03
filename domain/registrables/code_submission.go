package registrables

import (
	"context"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/google/uuid"
)

type codeSubmission struct {
	repo        repositories.CodeSubmissionRepository
	codeMatRepo repositories.CodeMaterialRepository
}

func NewCodeSubmission(repo repositories.CodeSubmissionRepository, codeMatRepo repositories.CodeMaterialRepository) registries.SubmissionRegistrable {
	return &codeSubmission{
		repo:        repo,
		codeMatRepo: codeMatRepo,
	}
}

type createCodeSubmissionPayload struct {
	Files    models.SubmissionFiles `json:"files"`
	RunnerID string                 `json:"runner_id"`
}

type updateCodeSubmissionPayload struct {
	Code           string                      `json:"code"`
	Status         string                      `json:"status"`
	AvgWallTime    float32                     `json:"avg_wall_time"`
	AvgMemory      int32                       `json:"avg_memory"`
	TestCaseGroups models.TestCaseGroupResults `json:"test_case_groups"`
}

func (c *codeSubmission) Create(ctx context.Context, uowRepo repositories.UoWInstance, submissionID string, matId string, payload []byte) error {
	parsedPayload, err := parsePayload[createCodeSubmissionPayload](payload)
	if err != nil {
		return errors.New("invalid payload type")
	}

	createPayload := &repositories.CreateCodeSubmissionPayload{
		SubmissionID: submissionID,
		Files:        parsedPayload.Files,
		RunnerID:     parsedPayload.RunnerID,
	}

	err = uowRepo.CodeSubmission().Create(ctx, createPayload)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	codeMat, err := c.codeMatRepo.GetByID(ctx, matId)
	if err != nil {
		return err
	}

	gradePayload := &models.GradeExecution{
		ID:       id.String(),
		Files:    parsedPayload.Files,
		RunnerID: parsedPayload.RunnerID,
		TaskID:   codeMat.TaskID,
	}

	return uowRepo.CodeSubmissionOutbox().Create(ctx, id.String(), submissionID, gradePayload)
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
