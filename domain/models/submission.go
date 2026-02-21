package models

import "time"

type SubmissionStatus string

const (
	QUEUED        SubmissionStatus = "queued"
	RUNNING       SubmissionStatus = "running"
	PASSED        SubmissionStatus = "passed"
	FAILED        SubmissionStatus = "failed"
	NOT_SUBMITTED SubmissionStatus = "not_submitted"
)

type Submission struct {
	ID        string           `json:"id"`
	Status    SubmissionStatus `json:"status"`
	Order     int              `json:"order"`
	CreatedAt time.Time        `json:"created_at"`
	Payload   any              `json:"payload"`
}

type RawSubmission struct {
	ID         string           `json:"id"`
	UserID     string           `json:"user_id"`
	MaterialID string           `json:"material_id"`
	LabID      string           `json:"lab_id"`
	SectionID  string           `json:"section_id"`
	CourseID   string           `json:"course_id"`
	Status     SubmissionStatus `json:"status"`
	Order      int              `json:"order"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Payload    any              `json:"payload"`

	ManualScore int    `json:"manual_score"`
	AutoScore   int    `json:"auto_score"`
	IPAddress   string `json:"ip_address"`
}

type SubmissionScore struct {
	Auto   int `json:"auto"`
	Manual int `json:"manual"`
}

type StudentLatestSubmission struct {
	ID        string          `json:"id"`
	User      Student         `json:"user"`
	Score     SubmissionScore `json:"score"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	IP        string          `json:"ip"`
	Payload   any             `json:"payload"`
}

type CMSSectionStudentSubmission struct {
	Student          Student          `json:"student"`
	AutoScore        int              `json:"auto_score"`
	ManualScore      int              `json:"manual_score"`
	IP               *string          `json:"ip"`
	CreatedAt        time.Time        `json:"created_at"`
	SubmissionStatus SubmissionStatus `json:"submission_status"`
	Payload          any              `json:"payload"`
}

type StudentSubmission struct {
	ID          string           `json:"id"`
	Status      SubmissionStatus `json:"status"`
	Order       int              `json:"order"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	AutoScore   int              `json:"auto_score"`
	ManualScore int              `json:"manual_score"`
	IP          string           `json:"ip"`
	Payload     any              `json:"payload"`
}
