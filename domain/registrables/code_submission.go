package registrables

import (
	"context"
	"errors"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/google/uuid"
)

// stripHiddenSegmentText removes every hidden segment's content from the given
// files. Used as a fail-closed fallback for submissions saved before
// StudentFiles existed: we cannot reconstruct the exact student view, but we can
// guarantee hidden grader code is never served by deleting any known hidden text.
// It collects hidden contents across all allowed runners (the submission's runner
// may be unknown on old rows) and removes every occurrence — over-removal is
// acceptable for legacy rows; a leak is not. Tasks with no hidden segments are
// returned unchanged.
// testCaseVisibility holds, per test case ID, whether the input and/or output
// must be withheld from students, plus the authoritative expected output to show
// in its place. Sourced from the task definition: a creator may hide the input,
// the output, or both, independently (CS-233).
type testCaseVisibility struct {
	hideInput      map[string]bool
	hideOutput     map[string]bool
	expectedOutput map[string]string
}

func testCaseVisibilityByID(task *taskPB.TaskResponse) testCaseVisibility {
	v := testCaseVisibility{
		hideInput:      make(map[string]bool),
		hideOutput:     make(map[string]bool),
		expectedOutput: make(map[string]string),
	}
	for _, tg := range task.GetTestCaseGroups() {
		for _, tc := range tg.GetTestCases() {
			if tc.GetHideInput() {
				v.hideInput[tc.GetId()] = true
			}
			if tc.GetHideOutput() {
				v.hideOutput[tc.GetId()] = true
			}
			v.expectedOutput[tc.GetId()] = tc.GetOutput()
		}
	}
	return v
}

// studentTestCaseResult projects a stored grader result into the student-facing
// view, honoring the creator's per-test-case hide flags. The output column shows
// the task's expected output (never the student's raw stdout, which could echo a
// hidden input).
//
// The message depends on status: on RUN_FAILED it is the compare script's
// output — instructor-authored feedback meant for the student, so it is always
// surfaced regardless of the hide flags. On every other non-passing status
// (TLE/MLE/runtime/signal/compile/grader) the message is raw stdout or grader
// output, which can leak a hidden input, so it is withheld whenever input or
// output is hidden.
func (v testCaseVisibility) studentTestCaseResult(tc models.TestCaseResult) models.TestCaseResult {
	result := models.TestCaseResult{
		ID:       tc.ID,
		Status:   tc.Status,
		WallTime: tc.WallTime,
		Memory:   tc.Memory,
	}
	if !v.hideInput[tc.ID] {
		result.Input = tc.Input
	}
	if !v.hideOutput[tc.ID] {
		result.Output = v.expectedOutput[tc.ID]
	}
	if tc.Status == models.CODE_EXECUTION_RUN_FAILED {
		// Compare-script output — student-facing feedback, safe even when hidden.
		result.Message = tc.Message
	} else if !v.hideInput[tc.ID] && !v.hideOutput[tc.ID] {
		result.Message = tc.Message
	}
	return result
}

func stripHiddenSegmentText(files models.SubmissionFiles, task *taskPB.TaskResponse) models.SubmissionFiles {
	var hidden []string
	for _, ar := range task.GetAllowedRunners() {
		for _, f := range ar.GetFiles() {
			for _, s := range f.GetSegments() {
				if s.GetType() == "hidden" && s.GetContent() != "" {
					hidden = append(hidden, s.GetContent())
				}
			}
		}
	}
	if len(hidden) == 0 {
		return files
	}

	out := make(models.SubmissionFiles, 0, len(files))
	for _, f := range files {
		content := f.Content
		for _, h := range hidden {
			content = strings.ReplaceAll(content, h, "")
		}
		out = append(out, models.SubmissionFile{Name: f.Name, Content: content})
	}
	return out
}

const defaultCompareScriptIDKey = "default_compare_script_id"

type CodeSubmission struct {
	repo            repositories.CodeSubmissionRepository
	codeMatRepo     repositories.CodeMaterialRepository
	submissionRepo  repositories.SubmissionRepository
	taskGRPCClient  taskPB.TaskServiceClient
	settingsRepo    repositories.SystemSettingsRepository
}

func NewCodeSubmission(repo repositories.CodeSubmissionRepository, codeMatRepo repositories.CodeMaterialRepository, submissionRepo repositories.SubmissionRepository, taskGRPCClient taskPB.TaskServiceClient, settingsRepo repositories.SystemSettingsRepository) *CodeSubmission {
	return &CodeSubmission{
		repo:           repo,
		codeMatRepo:    codeMatRepo,
		submissionRepo: submissionRepo,
		taskGRPCClient: taskGRPCClient,
		settingsRepo:   settingsRepo,
	}
}

type editableSegment struct {
	Index   int    `json:"index"`
	Content string `json:"content"`
}

type submittedFile struct {
	Name             string            `json:"name"`
	EditableSegments []editableSegment `json:"editable_segments"`
}

type createCodeSubmissionPayload struct {
	Files    []submittedFile `json:"files"`
	RunnerID string          `json:"runner_id"`
}

type updateCodeSubmissionPayload struct {
	Code           string                      `json:"code"`
	Status         string                      `json:"status"`
	AvgWallTime    float32                     `json:"avg_wall_time"`
	AvgMemory      int32                       `json:"avg_memory"`
	TestCaseGroups models.TestCaseGroupResults `json:"test_case_groups"`
}

type coreTestCaseGroupResult struct {
	ID      string                `json:"id"`
	Results []models.TestCaseResult `json:"results"`
}

type coreCodeSubmission struct {
	SubmissionID   string                    `json:"submission_id"`
	Files          models.SubmissionFiles    `json:"files"`
	RunnerID       *string                   `json:"runner_id"`
	Status         *string                   `json:"status"`
	AvgWallTime    *float32                  `json:"avg_wall_time"`
	AvgMemory      *int32                    `json:"avg_memory"`
	TestCaseGroups []coreTestCaseGroupResult `json:"test_case_groups"`
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

	codeMat, err := c.codeMatRepo.GetByID(ctx, matId)
	if err != nil {
		return err
	}

	task, err := c.taskGRPCClient.GetTask(ctx, &taskPB.GetTaskRequest{Id: codeMat.TaskID})
	if err != nil {
		return err
	}

	assembledFiles, err := assembleGraderFiles(parsedPayload.Files, parsedPayload.RunnerID, task)
	if err != nil {
		return err
	}

	// Student-facing files: same assembly minus hidden segments. Stored so the
	// student submission view never exposes hidden grader code (filtered in the
	// backend, deterministically, not reconstructed at read time).
	studentFiles, err := assembleStudentFiles(parsedPayload.Files, parsedPayload.RunnerID, task)
	if err != nil {
		return err
	}

	// Persist the raw indexed editable segments alongside the student files so the
	// segmented editor view can be rebuilt deterministically on read (no flat
	// reconstruction). These are the student's editable input only — never hidden
	// grader code — so they are safe to store on the student-facing files.
	attachEditableSegments(studentFiles, parsedPayload.Files)

	createPayload := &repositories.CreateCodeSubmissionPayload{
		SubmissionID: submissionID,
		Files:        assembledFiles,
		StudentFiles: studentFiles,
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

	gradePayload := &models.GradeExecution{
		ID:              id.String(),
		Files:           assembledFiles,
		RunnerID:        parsedPayload.RunnerID,
		TaskID:          codeMat.TaskID,
		CompareScriptID: c.resolveDefaultCompareScriptID(ctx),
	}

	return uowRepo.CodeSubmissionOutbox().Create(ctx, id.String(), submissionID, gradePayload)
}

// assembleGraderFiles builds the files the grader compiles and runs: editable
// (student) + readonly + hidden, in order. Exclude segments are dropped.
// Backward compat: if a task file has no segments, editable_segments[0].content
// is used as the full content.
func assembleGraderFiles(submitted []submittedFile, runnerID string, task *taskPB.TaskResponse) (models.SubmissionFiles, error) {
	return assembleFiles(submitted, runnerID, task, true /* includeHidden */, false /* includeExclude */)
}

// assembleStudentFiles builds the student-facing view of a submission: editable
// (student) + readonly + exclude, in order. Hidden segments are dropped so the
// student never sees the hidden grader code. Mirrors the editor display.
func assembleStudentFiles(submitted []submittedFile, runnerID string, task *taskPB.TaskResponse) (models.SubmissionFiles, error) {
	return assembleFiles(submitted, runnerID, task, false /* includeHidden */, true /* includeExclude */)
}

// assembleFiles assembles submission files from student editable segments and the
// task's segments. readonly is always kept; hidden/exclude are kept per the flags.
func assembleFiles(submitted []submittedFile, runnerID string, task *taskPB.TaskResponse, includeHidden, includeExclude bool) (models.SubmissionFiles, error) {
	// Find the task file map for the submitted runner.
	var taskFiles []*taskPB.File
	for _, ar := range task.GetAllowedRunners() {
		if ar.GetRunnerId() == runnerID {
			taskFiles = ar.GetFiles()
			break
		}
	}

	taskFileByName := make(map[string]*taskPB.File, len(taskFiles))
	for _, f := range taskFiles {
		taskFileByName[f.GetName()] = f
	}

	result := make(models.SubmissionFiles, 0, len(submitted))
	for _, sf := range submitted {
		taskFile, exists := taskFileByName[sf.Name]

		if !exists || len(taskFile.GetSegments()) == 0 {
			// Backward compat: no segments → use full content from first editable segment.
			content := ""
			if len(sf.EditableSegments) > 0 {
				content = sf.EditableSegments[0].Content
			}
			result = append(result, models.SubmissionFile{
				Name:    sf.Name,
				Content: content,
			})
			continue
		}

		// Build editable content map keyed by segment index.
		editableByIndex := make(map[int]string, len(sf.EditableSegments))
		for _, es := range sf.EditableSegments {
			editableByIndex[es.Index] = es.Content
		}

		// Assemble: iterate segments in order.
		assembled := ""
		for i, seg := range taskFile.GetSegments() {
			switch seg.GetType() {
			case "editable":
				assembled += editableByIndex[i]
			case "readonly":
				assembled += seg.GetContent()
			case "hidden":
				if includeHidden {
					assembled += seg.GetContent()
				}
			case "exclude":
				if includeExclude {
					assembled += seg.GetContent()
				}
			}
		}

		result = append(result, models.SubmissionFile{
			Name:    sf.Name,
			Content: assembled,
		})
	}

	return result, nil
}

// attachEditableSegments copies each submitted file's indexed editable segments
// onto the matching assembled student file (by name), in place. Used so the
// stored student files carry enough to rebuild the segmented editor view.
func attachEditableSegments(studentFiles models.SubmissionFiles, submitted []submittedFile) {
	segmentsByName := make(map[string][]editableSegment, len(submitted))
	for _, sf := range submitted {
		segmentsByName[sf.Name] = sf.EditableSegments
	}
	for i := range studentFiles {
		segs, ok := segmentsByName[studentFiles[i].Name]
		if !ok {
			continue
		}
		modelSegs := make([]models.EditableSegment, 0, len(segs))
		for _, s := range segs {
			modelSegs = append(modelSegs, models.EditableSegment{Index: s.Index, Content: s.Content})
		}
		studentFiles[i].EditableSegments = modelSegs
	}
}

func (c *CodeSubmission) resolveDefaultCompareScriptID(ctx context.Context) string {
	val, err := c.settingsRepo.Get(ctx, defaultCompareScriptIDKey)
	if err != nil || val == nil {
		return ""
	}
	return *val
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

	if viewBy != "instructor" {
		// Students must not see hidden grader code. Serve the stored student-facing
		// files (hidden dropped). For submissions saved before StudentFiles existed,
		// fail closed: strip any known hidden text from the full files rather than
		// risk leaking it.
		studentFiles := codeSubmission.StudentFiles
		if len(studentFiles) == 0 {
			studentFiles = stripHiddenSegmentText(codeSubmission.Files, task)
		}

		visibility := testCaseVisibilityByID(task)

		coreGroups := make([]coreTestCaseGroupResult, 0, len(testCaseGroups))
		for _, tg := range testCaseGroups {
			cg := coreTestCaseGroupResult{
				ID:      tg.ID,
				Results: make([]models.TestCaseResult, 0, len(tg.Results)),
			}
			for _, tc := range tg.Results {
				cg.Results = append(cg.Results, visibility.studentTestCaseResult(tc))
			}
			coreGroups = append(coreGroups, cg)
		}
		return &coreCodeSubmission{
			SubmissionID:   cleanedCodeSubmission.SubmissionID,
			Files:          studentFiles,
			RunnerID:       codeSubmission.RunnerID,
			Status:         cleanedCodeSubmission.Status,
			AvgWallTime:    cleanedCodeSubmission.AvgWallTime,
			AvgMemory:      cleanedCodeSubmission.AvgMemory,
			TestCaseGroups: coreGroups,
		}, nil
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

		if viewBy != "instructor" {
			// Students must not see hidden grader code — serve student-facing files.
			// Pre-StudentFiles submissions: fail closed by stripping known hidden text.
			studentFiles := codeSubmission.StudentFiles
			if len(studentFiles) == 0 {
				studentFiles = stripHiddenSegmentText(codeSubmission.Files, task)
			}

			visibility := testCaseVisibilityByID(task)

			coreGroups := make([]coreTestCaseGroupResult, 0, len(testCaseGroups))
			for _, tg := range testCaseGroups {
				cg := coreTestCaseGroupResult{
					ID:      tg.ID,
					Results: make([]models.TestCaseResult, 0, len(tg.Results)),
				}
				for _, tc := range tg.Results {
					cg.Results = append(cg.Results, visibility.studentTestCaseResult(tc))
				}
				coreGroups = append(coreGroups, cg)
			}
			submissions[subID] = &coreCodeSubmission{
				SubmissionID:   cleanedCodeSubmission.SubmissionID,
				Files:          studentFiles,
				Status:         cleanedCodeSubmission.Status,
				AvgWallTime:    cleanedCodeSubmission.AvgWallTime,
				AvgMemory:      cleanedCodeSubmission.AvgMemory,
				TestCaseGroups: coreGroups,
			}
		} else {
			submissions[subID] = cleanedCodeSubmission
		}
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
