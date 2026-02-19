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
	repo           repositories.CodeSubmissionRepository
	codeMatRepo    repositories.CodeMaterialRepository
	submissionRepo repositories.SubmissionRepository
}

func NewCodeSubmission(repo repositories.CodeSubmissionRepository, codeMatRepo repositories.CodeMaterialRepository, submissionRepo repositories.SubmissionRepository) registries.SubmissionRegistrable {
	return &codeSubmission{
		repo:           repo,
		codeMatRepo:    codeMatRepo,
		submissionRepo: submissionRepo,
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

func (c *codeSubmission) Get(ctx context.Context, submissionID string, viewBy string) (any, error) {
	submission, err := c.submissionRepo.Get(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	codeMat, err := c.codeMatRepo.GetByID(ctx, submission.MaterialID)
	if err != nil {
		return nil, err
	}

	codeSubmission, err := c.repo.Get(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	cleanedCodeSubmission := &models.CodeSubmission{
		SubmissionID:   codeSubmission.SubmissionID,
		Files:          codeSubmission.Files,
		Status:         codeSubmission.Status,
		AvgWallTime:    codeSubmission.AvgWallTime,
		AvgMemory:      codeSubmission.AvgMemory,
		TestCaseGroups: codeSubmission.TestCaseGroups,
	}

	if codeMat.HideTestCases && viewBy != "instructor" {
		testCaseGroups := make([]models.TestCaseGroupResult, 0, len(codeSubmission.TestCaseGroups))
		for _, tg := range codeSubmission.TestCaseGroups {
			_tg := models.TestCaseGroupResult{
				ID:      tg.ID,
				Score:   tg.Score,
				Results: make([]models.TestCaseResult, 0, len(tg.Results)),
			}

			for _, tc := range tg.Results {
				_tg.Results = append(_tg.Results, models.TestCaseResult{
					ID:       tc.ID,
					Message:  tc.Message,
					Status:   tc.Status,
					WallTime: tc.WallTime,
					Memory:   tc.Memory,
				})
			}
			testCaseGroups = append(testCaseGroups, _tg)
		}
		cleanedCodeSubmission.TestCaseGroups = testCaseGroups
	}

	return cleanedCodeSubmission, nil
}

func (c *codeSubmission) GetOverviewStats(payload any) any {
	codeSubmission, ok := payload.(*models.CodeSubmission)
	if !ok || codeSubmission == nil {
		return nil
	}

	passed, total := 0, 0
	for _, group := range codeSubmission.TestCaseGroups {
		total += len(group.Results)
		for _, tc := range group.Results {
			if tc.Status == models.CODE_EXECUTION_RUN_PASSED {
				passed++
			}
		}
	}

	return models.CodeSubmissionOverviewPayload{
		TotalTestCases:  total,
		PassedTestCases: passed,
	}
}

func (c *codeSubmission) GetOverviewStatsByID(ctx context.Context, submissionID string) any {
	codeSubmission, err := c.repo.Get(ctx, submissionID)
	if err != nil {
		return nil
	}

	passed, total := 0, 0
	for _, group := range codeSubmission.TestCaseGroups {
		total += len(group.Results)
		for _, tc := range group.Results {
			if tc.Status == models.CODE_EXECUTION_RUN_PASSED {
				passed++
			}
		}
	}

	return models.CodeSubmissionOverviewPayload{
		TotalTestCases:  total,
		PassedTestCases: passed,
	}
}

func (c *codeSubmission) GetOverviewsPayload(ctx context.Context, submissionIDs []string) (map[string]any, error) {
	if len(submissionIDs) == 0 {
		return map[string]any{}, nil
	}

	codeSubmissions, err := c.repo.GetByIDs(ctx, submissionIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any, len(codeSubmissions))
	for _, cs := range codeSubmissions {
		result[cs.SubmissionID] = c.GetOverviewStats(&models.CodeSubmission{
			TestCaseGroups: cs.TestCaseGroups,
		})
	}

	return result, nil
}
