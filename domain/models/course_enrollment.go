package models

import "time"

type CourseEnrollment struct {
	CourseID  string    `json:"course_id"`
	StudentID string    `json:"student_id"`
	CreatedAt time.Time `json:"created_at"`
}
