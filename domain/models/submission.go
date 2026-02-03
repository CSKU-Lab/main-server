package models

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
	Type    string           `json:"type"`
	Payload any              `json:"payload"`
}
