package registrables

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type codeMaterial struct {
	repo             repositories.CodeMaterialRepository
	taskGRPCClient   taskPB.TaskServiceClient
	configGRPCCLient configPB.ConfigServiceClient
}

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type CodeMaterialPayload struct {
	Description      *string     `json:"description,omitempty"`
	Solution         *string     `json:"solution,omitempty"`
	TestCases        *[]TestCase `json:"test_cases,omitempty"`
	AllowedRunnerIDs []string    `json:"allowed_runner_ids,omitempty"`
	CompareScriptID  *string     `json:"compare_script_id,omitempty"`
}

type Runner struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RunScript   string `json:"run_script"`
	BuildScript string `json:"build_script"`
}

type CompareScript struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CodeMaterialResponse struct {
	Description    string        `json:"description"`
	Solution       string        `json:"solution"`
	TestCases      []TestCase    `json:"test_cases"`
	AllowedRunners []Runner      `json:"allowed_runners"`
	CompareScript  CompareScript `json:"compare_script"`
}

func NewCodeMaterial(repo repositories.CodeMaterialRepository, taskGRPCClient taskPB.TaskServiceClient, configGRPCClient configPB.ConfigServiceClient) registries.MaterialRegisterable {
	return &codeMaterial{
		repo:             repo,
		taskGRPCClient:   taskGRPCClient,
		configGRPCCLient: configGRPCClient,
	}
}

func parsePayload(payload []byte) (*CodeMaterialPayload, error) {
	var matPayload struct {
		Payload CodeMaterialPayload `json:"payload"`
	}
	err := json.Unmarshal(payload, &matPayload)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			Message:    "invalid payload format",
			HttpStatus: http.StatusBadRequest,
		})
	}

	return &matPayload.Payload, nil
}

func (c *codeMaterial) GetByID(ctx context.Context, ID string) (any, error) {
	codeMat, err := c.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	task, err := c.taskGRPCClient.GetTask(ctx, &taskPB.GetTaskRequest{
		Id: codeMat.TaskID,
	})
	if err != nil {
		return nil, err
	}

	res := &CodeMaterialResponse{
		Description: codeMat.Description,
		Solution:    task.GetSolution(),
		TestCases:   make([]TestCase, len(task.GetTestcases())),
	}

	allowedRunners := make([]Runner, len(task.AllowedRunnerIds))
	for i, runnerID := range task.AllowedRunnerIds {
		runner, err := c.configGRPCCLient.GetRunner(ctx, &configPB.GetRunnerRequest{
			Id: runnerID,
		})
		if err != nil {
			return nil, err
		}

		allowedRunners[i] = Runner{
			ID:          runner.GetId(),
			Name:        runner.GetName(),
			BuildScript: runner.GetBuildScript(),
			RunScript:   runner.GetRunScript(),
		}
	}
	res.AllowedRunners = allowedRunners

	if task.CompareScriptId != "" {
		script, err := c.configGRPCCLient.GetCompare(ctx, &configPB.GetCompareRequest{
			Id: task.CompareScriptId,
		})
		if err != nil {
			return nil, err
		}

		compareScript := CompareScript{
			ID:   script.GetId(),
			Name: script.GetName(),
		}
		res.CompareScript = compareScript
	}

	for i, tc := range task.GetTestcases() {
		res.TestCases[i] = TestCase{
			Input:  tc.GetInput(),
			Output: tc.GetOutput(),
		}
	}

	return res, nil
}

func (c *codeMaterial) Create(ctx context.Context, matID string, req *requests.CreateMaterial, rawReq []byte) error {
	res, err := c.taskGRPCClient.UpsertTask(ctx, &taskPB.UpsertTaskRequest{})
	if err != nil {
		return err
	}

	if res.GetId() != "" {
		err = c.repo.SetTaskID(ctx, matID, res.GetId())
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *codeMaterial) UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial, rawReq []byte) error {
	payload, err := parsePayload(rawReq)
	if err != nil {
		return err
	}

	codeMat, err := c.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	if payload.Description != nil {
		err = c.repo.SetDescription(ctx, ID, *payload.Description)
		if err != nil {
			return err
		}
	}

	if payload.Solution != nil || payload.TestCases != nil || payload.AllowedRunnerIDs != nil || payload.CompareScriptID != nil {
		var testCases []*taskPB.UpsertTestCase
		if payload.TestCases != nil {
			testCases = make([]*taskPB.UpsertTestCase, 0, len(*payload.TestCases))
			for _, tc := range *payload.TestCases {
				testCases = append(testCases, &taskPB.UpsertTestCase{
					Input:  tc.Input,
					Output: tc.Output,
				})
			}
		}

		_, err = c.taskGRPCClient.UpsertTask(ctx, &taskPB.UpsertTaskRequest{
			Id:               &codeMat.TaskID,
			Solution:         payload.Solution,
			Testcases:        testCases,
			AllowedRunnerIds: payload.AllowedRunnerIDs,
			CompareId:        payload.CompareScriptID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *codeMaterial) DeleteByID(ctx context.Context, ID string) error {
	return nil
}
