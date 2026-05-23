package models

import (
	"time"
)

type Lab struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	CourseID    string    `json:"course_id"`
	IsDefault   bool      `json:"is_default"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CoreLabResponse struct {
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	ReadonlyAt    *time.Time `json:"readonly_at"`
	StudentStatus string     `json:"student_status"`
}
