package models

type SubmissionStatus string

const (
	QUEUED SubmissionStatus = "QUEUED"
	PASSED SubmissionStatus = "PASSED"
	FAILED SubmissionStatus = "FAILED"
)

type Submission struct {
	ID      string           `json:"id"`
	Status  SubmissionStatus `json:"status"`
	Type    string           `json:"type"`
	Payload any              `json:"payload"`
}
