package models

import (
	"time"
)

type DefaultLab struct {
	ID        string    `json:"id"`
	CourseID  string    `json:"course_id"`
	LabID     string    `json:"lab_id"`
	LabName   string    `json:"lab_name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
