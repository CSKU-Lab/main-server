package models

import (
	"time"
)

type Semester struct {
	ID        string       `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	Type      SemesterType `json:"type" db:"type"`
	StartDate time.Time    `json:"started_date" db:"started_date"`
}
