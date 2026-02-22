package models

import "time"

type MaterialStatus struct {
	Status      SubmissionStatus `json:"status"`
	SubmittedAt *time.Time       `json:"submitted_at"`
}

type MaterialCol struct {
	MaterialID   string `json:"material_id"`
	MaterialName string `json:"material_name"`
}

type StudentLabStatus struct {
	*Student
	MaterialStatuses map[string]MaterialStatus `json:"material_statuses"`
}

type LabStudentStatus struct {
	StudentRows  []StudentLabStatus `json:"student_rows"`
	MaterialCols []MaterialCol      `json:"material_cols"`
}
