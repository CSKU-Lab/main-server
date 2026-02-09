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
	ID        string           `json:"id"`
	Status    SubmissionStatus `json:"status"`
	Order     int              `json:"order"`
	CreatedAt time.Time        `json:"created_at"`
	Payload   any              `json:"payload"`
}
