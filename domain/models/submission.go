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
	UserID  string           `json:"user_id"`
	Status  SubmissionStatus `json:"status"`
	Payload any              `json:"payload"`
}
