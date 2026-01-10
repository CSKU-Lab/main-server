package registrables

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	graderPB "github.com/CSKU-Lab/main-server/genproto/grader/v1"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type codeMaterial struct {
	repo             repositories.CodeMaterialRepository
	taskGRPCClient   taskPB.TaskServiceClient
	configGRPCCLient configPB.ConfigServiceClient
	graderGRPCClient graderPB.GraderServiceClient
}

type TestCase struct {
	Order  int32  `json:"order"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type File struct {
	Name    string
	Content string
}

type Limit struct {
	CpuTime      float32 `json:"cpu_time"`
	CpuExtraTime float32 `json:"cpu_extra_time"`
	WallTime     float32 `json:"wall_time"`
	Memory       int32   `json:"memory"`
	Stack        int32   `json:"stack"`
	MaxOpenFiles int32   `json:"max_open_files"`
	MaxFileSize  float32 `json:"max_file_size"`
	NetworkAllow bool    `json:"network_allow"`
}

type CodeMaterialPayload struct {
	Description      *string     `json:"description,omitempty"`
	TestCases        *[]TestCase `json:"test_cases,omitempty"`
	AllowedRunnerIDs []string    `json:"allowed_runner_ids,omitempty"`
	CompareScriptID  *string     `json:"compare_script_id,omitempty"`
	SolutionRunnerID *string     `json:"solution_runner_id,omitempty"`
	SolutionFiles    []File      `json:"solution_files,omitempty"`
	Limit            *Limit      `json:"limit,omitempty"`
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
	Description    *string       `json:"description"`
	SolutionFiles  []File        `json:"solution_files"`
	TestCases      []TestCase    `json:"test_cases"`
	AllowedRunners []Runner      `json:"allowed_runners"`
	CompareScript  CompareScript `json:"compare_script"`
}

func NewCodeMaterial(repo repositories.CodeMaterialRepository, taskGRPCClient taskPB.TaskServiceClient, configGRPCClient configPB.ConfigServiceClient, graderGRPCClient graderPB.GraderServiceClient) registries.MaterialRegisterable {
	return &codeMaterial{
		repo:             repo,
		taskGRPCClient:   taskGRPCClient,
		configGRPCCLient: configGRPCClient,
		graderGRPCClient: graderGRPCClient,
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

	solutionFilesRes := make([]File, 0, len(task.GetSolutionFiles()))
	for _, f := range task.GetSolutionFiles() {
		solutionFilesRes = append(solutionFilesRes, File{
			Name:    f.GetName(),
			Content: f.GetContent(),
		})
	}

	res := &CodeMaterialResponse{
		Description:   codeMat.Description,
		TestCases:     make([]TestCase, len(task.GetTestcases())),
		SolutionFiles: solutionFilesRes,
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

	if task.CompareScriptId != nil {
		script, err := c.configGRPCCLient.GetCompare(ctx, &configPB.GetCompareRequest{
			Id: task.GetCompareScriptId(),
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
			Order:  tc.GetOrder(),
			Input:  tc.GetInput(),
			Output: tc.GetOutput(),
		}
	}

	return res, nil
}

func (c *codeMaterial) Create(ctx context.Context, matID string, req *requests.CreateMaterial, rawReq []byte) error {
	res, err := c.taskGRPCClient.CreateTask(ctx, nil)
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

	var testCases []*taskPB.TestCase
	if payload.TestCases != nil {
		testCases = make([]*taskPB.TestCase, 0, len(*payload.TestCases))
		for _, tc := range *payload.TestCases {
			testCases = append(testCases, &taskPB.TestCase{
				Order: tc.Order,
				Input: tc.Input,
			})
		}
	}

	taskPBSolutionFiles := make([]*taskPB.SolutionFile, 0, len(payload.SolutionFiles))
	for _, f := range payload.SolutionFiles {
		taskPBSolutionFiles = append(taskPBSolutionFiles, &taskPB.SolutionFile{
			Name:    f.Name,
			Content: f.Content,
		})
	}

	var limit *taskPB.Limit
	if payload.Limit != nil {
		limit = &taskPB.Limit{
			CpuTime:      payload.Limit.CpuTime,
			CpuExtraTime: payload.Limit.CpuExtraTime,
			WallTime:     payload.Limit.WallTime,
			Memory:       payload.Limit.Memory,
			Stack:        payload.Limit.Stack,
			MaxOpenFiles: payload.Limit.MaxOpenFiles,
			MaxFileSize:  payload.Limit.MaxFileSize,
			NetworkAllow: payload.Limit.NetworkAllow,
		}
	}

	_, err = c.taskGRPCClient.UpdateTask(ctx, &taskPB.UpdateTaskRequest{
		Id:               &codeMat.TaskID,
		Testcases:        testCases,
		AllowedRunnerIds: payload.AllowedRunnerIDs,
		SolutionFiles:    taskPBSolutionFiles,
		SolutionRunnerId: payload.SolutionRunnerID,
		CompareScriptId:  payload.CompareScriptID,
		Limit:            limit,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *codeMaterial) DeleteByID(ctx context.Context, ID string) error {
	return nil
}
