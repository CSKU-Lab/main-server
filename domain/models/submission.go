package models

import "time"

type SubmissionStatus string

const (
	QUEUED  SubmissionStatus = "queued"
	RUNNING SubmissionStatus = "running"
	PASSED  SubmissionStatus = "passed"
	FAILED  SubmissionStatus = "failed"
)

type Submission struct {
	ID      string           `json:"id"`
	Status  SubmissionStatus `json:"status"`
	Order   int              `json:"order"`
	Payload any              `json:"payload"`
}

type SubmissionOverview struct {
	ID        string           `json:"id"`
	Status    SubmissionStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	Payload   any              `json:"payload"`
}

type CodeSubmissionOverviewPayload struct {
	TotalTestCases  int `json:"total_test_cases"`
	PassedTestCases int `json:"passed_test_cases"`
}
