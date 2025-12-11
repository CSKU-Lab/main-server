package models

import (
	"time"
)

type Lab struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	CourseID    string    `json:"course_id"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
