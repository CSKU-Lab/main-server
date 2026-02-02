package models

type CodeSubmission struct {
	Code           string          `json:"code"`
	Status         *string         `json:"status"`
	AvgWallTime    *float32        `json:"avg_wall_time"`
	AvgMemory      *int32          `json:"avg_memory"`
	TestCaseGroups []TestCaseGroup `json:"test_case_groups"`
}
