package registrables

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CodeMaterial struct {
	repo             repositories.CodeMaterialRepository
	taskGRPCClient   taskPB.TaskServiceClient
	configGRPCCLient configPB.ConfigServiceClient
	settingsRepo     repositories.SystemSettingsRepository
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

type Limits struct {
	CpuTime      float32 `json:"cpu_time"`
	CpuExtraTime float32 `json:"cpu_extra_time"`
	WallTime     float32 `json:"wall_time"`
	Memory       int32   `json:"memory"`
	Stack        int32   `json:"stack"`
	MaxOpenFiles int32   `json:"max_open_files"`
	MaxFileSize  float32 `json:"max_file_size"`
	NetworkAllow bool    `json:"network_allow"`
}

// Solution groups the solution runner and its files together (request payload).
type Solution struct {
	RunnerID string `json:"runner_id"`
	Files    []File `json:"files"`
}

// AllowedRunner represents a runner with per-material customized starter files (request payload).
type AllowedRunner struct {
	RunnerID string `json:"runner_id"`
	Files    []File `json:"files"`
}

// RunnerInfo holds the basic identifying metadata for a runner.
type RunnerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SolutionResponse is the solution as returned in the CMS GET response.
type SolutionResponse struct {
	Runner RunnerInfo `json:"runner"`
	Files  []File     `json:"files"`
}

// AllowedRunnerResponse is a runner as returned in the CMS GET response.
type AllowedRunnerResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RunScript   string `json:"run_script"`
	BuildScript string `json:"build_script"`
	Files       []File `json:"files"`
}

type CompareScript struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CodeMaterialPayload is the request body under "payload" for PATCH /materials/:id.
type CodeMaterialPayload struct {
	Description     *string          `json:"description,omitempty"`
	TestCaseGroups  *[]TestCaseGroup `json:"test_case_groups,omitempty"`
	AllowedRunners  []AllowedRunner  `json:"allowed_runners,omitempty"`
	CompareScriptID *string          `json:"compare_script_id,omitempty"`
	Solution        *Solution        `json:"solution,omitempty"`
	ResourceFiles   []File           `json:"resource_files,omitempty"`
	Limits          *Limits          `json:"limits,omitempty"`
	HideTestCases   *bool            `json:"hide_test_cases"`
}

// CodeMaterialResponse is the full payload returned from GET /materials/:id (CMS).
type CodeMaterialResponse struct {
	Description    *string                 `json:"description"`
	Solution       *SolutionResponse       `json:"solution"`
	ResourceFiles  []File                  `json:"resource_files"`
	TestCaseGroups []TestCaseGroup         `json:"test_case_groups"`
	AllowedRunners []AllowedRunnerResponse `json:"allowed_runners"`
	CompareScript  *CompareScript          `json:"compare_script"`
	Limits         *Limits                 `json:"limits"`
	HideTestCases  bool                    `json:"hide_test_cases"`
}

func NewCodeMaterial(repo repositories.CodeMaterialRepository, taskGRPCClient taskPB.TaskServiceClient, configGRPCClient configPB.ConfigServiceClient, settingsRepo repositories.SystemSettingsRepository) *CodeMaterial {
	return &CodeMaterial{
		repo:             repo,
		taskGRPCClient:   taskGRPCClient,
		configGRPCCLient: configGRPCClient,
		settingsRepo:     settingsRepo,
	}
}

func (c *CodeMaterial) CalculateScores(rawReq []byte) (*registries.MaterialScores, error) {
	payload, err := parsePayload[CodeMaterialPayload](rawReq)
	if err != nil {
		return nil, err
	}

	if payload == nil || payload.TestCaseGroups == nil {
		return &registries.MaterialScores{
			AutoScore:   0,
			ManualScore: 0,
		}, nil
	}

	var autoScore int32
	var manualScore int32
	for _, g := range *payload.TestCaseGroups {
		autoScore += g.Score
	}

	return &registries.MaterialScores{
		AutoScore:   int(autoScore),
		ManualScore: int(manualScore),
	}, nil
}

func (c *CodeMaterial) GetByID(ctx context.Context, ID string) (any, error) {
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

	// Build resource files
	resourceFiles := make([]File, 0, len(task.GetResourceFiles()))
	for _, f := range task.GetResourceFiles() {
		resourceFiles = append(resourceFiles, File{
			Name:    f.GetName(),
			Content: f.GetContent(),
		})
	}

	// Build limit
	var limit *Limits
	if task.Limit != nil {
		limit = &Limits{
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

	// Build solution (runner metadata from config + files from task)
	var solution *SolutionResponse
	if task.GetSolution() != nil {
		pbSolution := task.GetSolution()
		runner, err := c.configGRPCCLient.GetRunner(ctx, &configPB.GetRunnerRequest{
			Id: pbSolution.GetRunnerId(),
		})
		if err != nil {
			return nil, err
		}

		solutionFiles := make([]File, 0, len(pbSolution.GetFiles()))
		for _, f := range pbSolution.GetFiles() {
			solutionFiles = append(solutionFiles, File{
				Name:    f.GetName(),
				Content: f.GetContent(),
			})
		}

		solution = &SolutionResponse{
			Runner: RunnerInfo{
				ID:   runner.GetId(),
				Name: runner.GetName(),
			},
			Files: solutionFiles,
		}
	}

	res := &CodeMaterialResponse{
		Description:    codeMat.Description,
		TestCaseGroups: make([]TestCaseGroup, len(task.GetTestCaseGroups())),
		Solution:       solution,
		ResourceFiles:  resourceFiles,
		Limits:         limit,
	}

	// Build allowed runners (runner metadata from config + custom files from task)
	allowedRunners := make([]AllowedRunnerResponse, len(task.GetAllowedRunners()))
	for i, pbRunner := range task.GetAllowedRunners() {
		runner, err := c.configGRPCCLient.GetRunner(ctx, &configPB.GetRunnerRequest{
			Id: pbRunner.GetRunnerId(),
		})
		if err != nil {
			return nil, err
		}

		runnerFiles := make([]File, 0, len(pbRunner.GetFiles()))
		for _, f := range pbRunner.GetFiles() {
			runnerFiles = append(runnerFiles, File{
				Name:    f.GetName(),
				Content: f.GetContent(),
			})
		}

		allowedRunners[i] = AllowedRunnerResponse{
			ID:          runner.GetId(),
			Name:        runner.GetName(),
			BuildScript: runner.GetBuildScript(),
			RunScript:   runner.GetRunScript(),
			Files:       runnerFiles,
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

func (c *CodeMaterial) Create(ctx context.Context, matID string, req *requests.CreateMaterial, rawReq []byte) error {
	res, err := c.taskGRPCClient.CreateTask(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	if res.GetId() == "" {
		return nil
	}

	if err = c.repo.SetTaskID(ctx, matID, res.GetId()); err != nil {
		return err
	}

	defaultCompare, err := c.settingsRepo.Get(ctx, defaultCompareScriptIDKey)
	if err == nil && defaultCompare != nil && *defaultCompare != "" {
		taskID := res.GetId()
		_, _ = c.taskGRPCClient.UpdateTask(ctx, &taskPB.UpdateTaskRequest{
			Id:              &taskID,
			CompareScriptId: defaultCompare,
		})
	}

	return nil
}

func (c *CodeMaterial) UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial, rawReq []byte) error {
	payload, err := parsePayload[CodeMaterialPayload](rawReq)
	if err != nil {
		return err
	}

	// no code material payload in the request
	if payload == nil {
		return nil
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

	// Build test case groups
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
				testCaseGroup.TestCases[i] = &taskPB.TestCase{
					Id:       tc.ID,
					Order:    tc.Order,
					Input:    tc.Input,
					Output:   tc.Output,
					IsHidden: tc.IsHidden,
				}
			}

			testCaseGroups = append(testCaseGroups, testCaseGroup)
		}
	}

	// Build allowed runners
	pbAllowedRunners := make([]*taskPB.AllowedRunner, 0, len(payload.AllowedRunners))
	for _, ar := range payload.AllowedRunners {
		pbFiles := make([]*taskPB.File, 0, len(ar.Files))
		for _, f := range ar.Files {
			pbFiles = append(pbFiles, &taskPB.File{
				Name:    f.Name,
				Content: f.Content,
			})
		}
		pbAllowedRunners = append(pbAllowedRunners, &taskPB.AllowedRunner{
			RunnerId: ar.RunnerID,
			Files:    pbFiles,
		})
	}

	// Build solution
	var pbSolution *taskPB.Solution
	if payload.Solution != nil {
		pbSolutionFiles := make([]*taskPB.File, 0, len(payload.Solution.Files))
		for _, f := range payload.Solution.Files {
			pbSolutionFiles = append(pbSolutionFiles, &taskPB.File{
				Name:    f.Name,
				Content: f.Content,
			})
		}
		pbSolution = &taskPB.Solution{
			RunnerId: payload.Solution.RunnerID,
			Files:    pbSolutionFiles,
		}
	}

	// Build resource files
	pbResourceFiles := make([]*taskPB.File, 0, len(payload.ResourceFiles))
	for _, f := range payload.ResourceFiles {
		pbResourceFiles = append(pbResourceFiles, &taskPB.File{
			Name:    f.Name,
			Content: f.Content,
		})
	}

	// Build limit
	var limit *taskPB.Limit
	if payload.Limits != nil {
		limit = &taskPB.Limit{
			CpuTime:      payload.Limits.CpuTime,
			CpuExtraTime: payload.Limits.CpuExtraTime,
			WallTime:     payload.Limits.WallTime,
			Memory:       payload.Limits.Memory,
			Stack:        payload.Limits.Stack,
			MaxOpenFiles: payload.Limits.MaxOpenFiles,
			MaxFileSize:  payload.Limits.MaxFileSize,
			NetworkAllow: payload.Limits.NetworkAllow,
		}
	}

	_, err = c.taskGRPCClient.UpdateTask(ctx, &taskPB.UpdateTaskRequest{
		Id:              &codeMat.TaskID,
		TestCaseGroups:  testCaseGroups,
		AllowedRunners:  pbAllowedRunners,
		Solution:        pbSolution,
		ResourceFiles:   pbResourceFiles,
		CompareScriptId: payload.CompareScriptID,
		Limit:           limit,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *CodeMaterial) DeleteByID(ctx context.Context, ID string) error {
	return nil
}

func (c *CodeMaterial) Clone(ctx context.Context, sourceID string, targetID string) error {
	source, err := c.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}

	sourcePayload := source.(*CodeMaterialResponse)
	if err := c.Create(ctx, targetID, nil, nil); err != nil {
		return err
	}

	testCaseGroups := sourcePayload.TestCaseGroups
	payload := &CodeMaterialPayload{
		Description:    sourcePayload.Description,
		TestCaseGroups: &testCaseGroups,
		ResourceFiles:  sourcePayload.ResourceFiles,
		Limits:         sourcePayload.Limits,
		HideTestCases:  &sourcePayload.HideTestCases,
	}

	payload.AllowedRunners = make([]AllowedRunner, 0, len(sourcePayload.AllowedRunners))
	for _, runner := range sourcePayload.AllowedRunners {
		payload.AllowedRunners = append(payload.AllowedRunners, AllowedRunner{
			RunnerID: runner.ID,
			Files:    runner.Files,
		})
	}

	if sourcePayload.Solution != nil {
		payload.Solution = &Solution{
			RunnerID: sourcePayload.Solution.Runner.ID,
			Files:    sourcePayload.Solution.Files,
		}
	}

	if sourcePayload.CompareScript != nil {
		payload.CompareScriptID = &sourcePayload.CompareScript.ID
	}

	rawReq, err := buildPayloadRequest(payload)
	if err != nil {
		return err
	}

	return c.UpdateByID(ctx, targetID, &requests.BaseUpdateMaterial{}, rawReq)
}
