package models

import (
	"time"
)

type Semester struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Type      SemesterType `json:"type"`
	StartDate time.Time    `json:"started_date"`
}
