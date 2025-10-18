package models

type AffectedSection struct {
	CourseName string   `json:"course_name"`
	Sections   []string `json:"sections"`
}
