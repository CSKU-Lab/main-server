package registrables

import (
	"context"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/google/uuid"
)

type CodeSubmission struct {
	repo           repositories.CodeSubmissionRepository
	codeMatRepo    repositories.CodeMaterialRepository
	submissionRepo repositories.SubmissionRepository
	taskGRPCClient taskPB.TaskServiceClient
}

func NewCodeSubmission(repo repositories.CodeSubmissionRepository, codeMatRepo repositories.CodeMaterialRepository, submissionRepo repositories.SubmissionRepository, taskGRPCClient taskPB.TaskServiceClient) *CodeSubmission {
	return &CodeSubmission{
		repo:           repo,
		codeMatRepo:    codeMatRepo,
		submissionRepo: submissionRepo,
		taskGRPCClient: taskGRPCClient,
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

func reorderTestCaseGroups(groups []models.TestCaseGroupResult, taskGroups []*taskPB.TestCaseGroup) []models.TestCaseGroupResult {
	resultGroupByID := make(map[string]map[string]models.TestCaseResult, len(groups))
	for _, g := range groups {
		resultGroupByID[g.ID] = make(map[string]models.TestCaseResult, len(g.Results))
		for _, tc := range g.Results {
			resultGroupByID[g.ID][tc.ID] = tc
		}
	}

	ordered := make([]models.TestCaseGroupResult, 0, len(taskGroups))
	for _, tg := range taskGroups {
		if _, exists := resultGroupByID[tg.GetId()]; !exists {
			continue
		}

		orderedResults := make([]models.TestCaseResult, 0, len(tg.GetTestCases()))
		for _, tc := range tg.GetTestCases() {
			if tcResult, ok := resultGroupByID[tg.GetId()][tc.GetId()]; ok {
				orderedResults = append(orderedResults, tcResult)
			}
		}

		ordered = append(ordered, models.TestCaseGroupResult{
			ID:      tg.GetId(),
			Score:   tg.GetScore(),
			Results: orderedResults,
		})
	}
	return ordered
}

func (c *CodeSubmission) Create(ctx context.Context, uowRepo repositories.UoWInstance, submissionID string, matId string, payload []byte) error {
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

func (c *CodeSubmission) Update(ctx context.Context, uowRepo repositories.UoWInstance, submissionID string, payload []byte) error {
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

func (c *CodeSubmission) Get(ctx context.Context, submissionID string, viewBy string) (any, error) {
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

	task, err := c.taskGRPCClient.GetTask(ctx, &taskPB.GetTaskRequest{
		Id: codeMat.TaskID,
	})
	if err != nil {
		return nil, err
	}

	testCaseGroups := codeSubmission.TestCaseGroups
	if len(task.GetTestCaseGroups()) > 0 {
		testCaseGroups = reorderTestCaseGroups(codeSubmission.TestCaseGroups, task.GetTestCaseGroups())
	}

	cleanedCodeSubmission := &models.CodeSubmission{
		SubmissionID:   codeSubmission.SubmissionID,
		Files:          codeSubmission.Files,
		Status:         codeSubmission.Status,
		AvgWallTime:    codeSubmission.AvgWallTime,
		AvgMemory:      codeSubmission.AvgMemory,
		TestCaseGroups: testCaseGroups,
	}

	if codeMat.HideTestCases && viewBy != "instructor" {
		cleanedTestCaseGroups := make([]models.TestCaseGroupResult, 0, len(testCaseGroups))
		for _, tg := range testCaseGroups {
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
			cleanedTestCaseGroups = append(cleanedTestCaseGroups, _tg)
		}
		cleanedCodeSubmission.TestCaseGroups = cleanedTestCaseGroups
	}

	return cleanedCodeSubmission, nil
}

func (c *CodeSubmission) GetByIDs(ctx context.Context, submissionIDs []string, viewBy string) (map[string]any, error) {
	submissions := make(map[string]any, len(submissionIDs))
	for _, subID := range submissionIDs {
		submission, err := c.submissionRepo.Get(ctx, subID)
		if err != nil {
			return nil, err
		}

		codeMat, err := c.codeMatRepo.GetByID(ctx, submission.MaterialID)
		if err != nil {
			return nil, err
		}

		codeSubmission, err := c.repo.Get(ctx, subID)
		if err != nil {
			return nil, err
		}

		task, err := c.taskGRPCClient.GetTask(ctx, &taskPB.GetTaskRequest{
			Id: codeMat.TaskID,
		})
		if err != nil {
			return nil, err
		}

		testCaseGroups := codeSubmission.TestCaseGroups
		if len(task.GetTestCaseGroups()) > 0 {
			testCaseGroups = reorderTestCaseGroups(codeSubmission.TestCaseGroups, task.GetTestCaseGroups())
		}

		cleanedCodeSubmission := &models.CodeSubmission{
			SubmissionID:   codeSubmission.SubmissionID,
			Files:          codeSubmission.Files,
			Status:         codeSubmission.Status,
			AvgWallTime:    codeSubmission.AvgWallTime,
			AvgMemory:      codeSubmission.AvgMemory,
			TestCaseGroups: testCaseGroups,
		}

		if codeMat.HideTestCases && viewBy != "instructor" {
			cleanedTestCaseGroups := make([]models.TestCaseGroupResult, 0, len(testCaseGroups))
			for _, tg := range testCaseGroups {
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
				cleanedTestCaseGroups = append(cleanedTestCaseGroups, _tg)
			}
			cleanedCodeSubmission.TestCaseGroups = cleanedTestCaseGroups
		}

		submissions[subID] = cleanedCodeSubmission
	}

	return submissions, nil
}

func (c *CodeSubmission) GetOverviewStats(payload any) any {
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

func (c *CodeSubmission) GetOverviewStatsByID(ctx context.Context, submissionID string) any {
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

func (c *CodeSubmission) GetOverviewsPayload(ctx context.Context, submissionIDs []string) (map[string]any, error) {
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
