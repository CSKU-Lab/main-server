package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type CodeSubmission struct {
	SubmissionID   string                `json:"submission_id"`
	Files          SubmissionFiles       `json:"files"`
	StudentFiles   SubmissionFiles       `json:"student_files"`
	Status         *string               `json:"status"`
	AvgWallTime    *float32              `json:"avg_wall_time"`
	AvgMemory      *int32                `json:"avg_memory"`
	TestCaseGroups []TestCaseGroupResult `json:"test_case_groups"`
	RunnerID       *string               `json:"runner_id"`
}

type EditableSegment struct {
	Index   int    `json:"index"`
	Content string `json:"content"`
}

type SubmissionFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	// EditableSegments carries the student's per-slot editable input, indexed to
	// the runner template segments. Persisted on student files only so a later
	// read can rebuild the segmented editor view (readonly ranges + editable
	// content) without ambiguous flat-content reconstruction. Empty for grader
	// files and for submissions saved before this field existed.
	EditableSegments []EditableSegment `json:"editable_segments,omitempty"`
}

type SubmissionFiles []SubmissionFile

func (t SubmissionFiles) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *SubmissionFiles) Scan(src any) error {
	var data []byte
	switch src := src.(type) {
	case []byte:
		data = src
	case string:
		data = []byte(src)
	case nil:
		return nil
	default:
		return errors.New("unsupported data type for SubmissionFiles")
	}

	return json.Unmarshal(data, t)
}

type GradeExecution struct {
	ID              string           `json:"id"`
	Files           []SubmissionFile `json:"files"`
	TaskID          string           `json:"task_id"`
	RunnerID        string           `json:"runner_id"`
	CompareScriptID string           `json:"compare_script_id,omitempty"`
}

type CodeExecutionStatus string

const (
	CODE_EXECUTION_COMPILE_FAILED        CodeExecutionStatus = "COMPILE_FAILED"
	CODE_EXECUTION_RUN_PASSED            CodeExecutionStatus = "RUN_PASSED"
	CODE_EXECUTION_RUN_FAILED            CodeExecutionStatus = "RUN_FAILED"
	CODE_EXECUTION_TIME_LIMIT_EXCEEDED   CodeExecutionStatus = "TIME_LIMIT_EXCEEDED"
	CODE_EXECUTION_MEMORY_LIMIT_EXCEEDED CodeExecutionStatus = "MEMORY_LIMIT_EXCEEDED"
	CODE_EXECUTION_RUNTIME_ERROR         CodeExecutionStatus = "RUNTIME_ERROR"
	CODE_EXECUTION_SIGNAL_ERROR          CodeExecutionStatus = "SIGNAL_ERROR"
	CODE_EXECUTION_GRADER_ERROR          CodeExecutionStatus = "GRADER_ERROR"
	CODE_EXECUTION_FILE_NOT_FOUND        CodeExecutionStatus = "FILE_NOT_FOUND"
	CODE_EXECUTION_BUILD_PASSED          CodeExecutionStatus = "BUILD_PASSED"
	CODE_EXECUTION_QUEUED                CodeExecutionStatus = "QUEUED"
	CODE_EXECUTION_RUNNING               CodeExecutionStatus = "RUNNING"
)

type TestCaseResult struct {
	ID       string              `json:"id"`
	Status   CodeExecutionStatus `json:"status"`
	Input    string              `json:"input"`
	Output   string              `json:"output"`
	Message  string              `json:"message"`
	WallTime float32             `json:"wall_time"`
	Memory   int32               `json:"memory"`
}

type TestCaseGroupResult struct {
	ID      string           `json:"id"`
	Score   int32            `json:"score"`
	Results []TestCaseResult `json:"results"`
}

type TestCaseGroupResults []TestCaseGroupResult

func (t TestCaseGroupResults) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TestCaseGroupResults) Scan(src any) error {
	var data []byte
	switch src := src.(type) {
	case []byte:
		data = src
	case string:
		data = []byte(src)
	case nil:
		return nil
	default:
		return errors.New("unsupported data type for TestCaseGroups")
	}

	return json.Unmarshal(data, t)
}

type GradeResult struct {
	Status               CodeExecutionStatus   `json:"status"`
	TestCaseGroupResults []TestCaseGroupResult `json:"test_case_group_results"`
	AvgWallTime          float32               `json:"avg_wall_time"`
	AvgMemory            int32                 `json:"avg_memory"`
	Score                int32                 `json:"score"`
}

type CodeSubmissionOverviewPayload struct {
	TotalTestCases  int `json:"total_test_cases"`
	PassedTestCases int `json:"passed_test_cases"`
}
