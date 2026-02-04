package registrables

import (
	"context"

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
	ID       string `json:"id"`
	Order    int32  `json:"order"`
	Input    string `json:"input"`
	Output   string `json:"output"`
	IsHidden bool   `json:"is_hidden"`
}

type TestCaseGroup struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Score     int32      `json:"score"`
	Order     int32      `json:"order"`
	TestCases []TestCase `json:"test_cases"`
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
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
	Description      *string          `json:"description,omitempty"`
	TestCaseGroups   *[]TestCaseGroup `json:"test_case_groups,omitempty"`
	AllowedRunnerIDs []string         `json:"allowed_runner_ids,omitempty"`
	CompareScriptID  *string          `json:"compare_script_id,omitempty"`
	SolutionRunnerID *string          `json:"solution_runner_id,omitempty"`
	SolutionFiles    []File           `json:"solution_files,omitempty"`
	Limit            *Limit           `json:"limit,omitempty"`
	HideTestCases    *bool            `json:"hide_test_cases"`
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
	Description      *string         `json:"description"`
	SolutionFiles    []File          `json:"solution_files"`
	SolutionRunnerID *string         `json:"solution_runner_id"`
	TestCaseGroups   []TestCaseGroup `json:"test_case_groups"`
	AllowedRunners   []Runner        `json:"allowed_runners"`
	CompareScript    *CompareScript  `json:"compare_script"`
	Limit            *Limit          `json:"limit"`
	HideTestCases    bool            `json:"hide_test_cases"`
}

func NewCodeMaterial(repo repositories.CodeMaterialRepository, taskGRPCClient taskPB.TaskServiceClient, configGRPCClient configPB.ConfigServiceClient, graderGRPCClient graderPB.GraderServiceClient) registries.MaterialRegisterable {
	return &codeMaterial{
		repo:             repo,
		taskGRPCClient:   taskGRPCClient,
		configGRPCCLient: configGRPCClient,
		graderGRPCClient: graderGRPCClient,
	}
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

	var limit *Limit
	if task.Limit != nil {
		limit = &Limit{
			CpuTime:      task.GetLimit().GetCpuTime(),
			CpuExtraTime: task.GetLimit().GetCpuExtraTime(),
			WallTime:     task.GetLimit().GetWallTime(),
			Memory:       task.GetLimit().GetMemory(),
			Stack:        task.GetLimit().GetStack(),
			MaxOpenFiles: task.GetLimit().GetMaxOpenFiles(),
			MaxFileSize:  task.GetLimit().GetMaxFileSize(),
			NetworkAllow: task.GetLimit().GetNetworkAllow(),
		}
	}

	res := &CodeMaterialResponse{
		Description:      codeMat.Description,
		TestCaseGroups:   make([]TestCaseGroup, len(task.GetTestCaseGroups())),
		SolutionFiles:    solutionFilesRes,
		SolutionRunnerID: task.SolutionRunnerId,
		Limit:            limit,
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

		res.CompareScript = &CompareScript{
			ID:   script.GetId(),
			Name: script.GetName(),
		}
	}

	for i, tc := range task.GetTestCaseGroups() {
		res.TestCaseGroups[i] = TestCaseGroup{
			ID:        tc.GetId(),
			Name:      tc.GetName(),
			Order:     tc.GetOrder(),
			Score:     tc.GetScore(),
			TestCases: make([]TestCase, len(tc.GetTestCases())),
		}

		for j, tcase := range tc.GetTestCases() {
			res.TestCaseGroups[i].TestCases[j] = TestCase{
				ID:       tcase.GetId(),
				Order:    tcase.GetOrder(),
				Input:    tcase.GetInput(),
				Output:   tcase.GetOutput(),
				IsHidden: tcase.GetIsHidden(),
			}
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
	payload, err := parsePayload[CodeMaterialPayload](rawReq)
	if err != nil {
		return err
	}

	codeMat, err := c.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	err = c.repo.Update(ctx, ID, &repositories.UpdateCodeMaterialPayload{
		Description:   payload.Description,
		HideTestCases: payload.HideTestCases,
	})
	if err != nil {
		return err
	}

	var testCaseGroups []*taskPB.TestCaseGroup
	if payload.TestCaseGroups != nil {
		for _, g := range *payload.TestCaseGroups {
			testCaseGroup := &taskPB.TestCaseGroup{
				Id:        g.ID,
				Name:      g.Name,
				Score:     g.Score,
				Order:     g.Order,
				TestCases: make([]*taskPB.TestCase, len(g.TestCases)),
			}

			for i, tc := range g.TestCases {
				testCase := &taskPB.TestCase{
					Id:       tc.ID,
					Order:    tc.Order,
					Input:    tc.Input,
					Output:   tc.Output,
					IsHidden: tc.IsHidden,
				}
				testCaseGroup.TestCases[i] = testCase
			}

			testCaseGroups = append(testCaseGroups, testCaseGroup)
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
		TestCaseGroups:   testCaseGroups,
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
