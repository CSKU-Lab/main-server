package models

import (
	"time"

	"github.com/CSKU-Lab/main-server/constants"
)

type Semester struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Type      constants.SemesterType `json:"type" db:"type"`
	StartDate time.Time              `json:"started_date" db:"started_date"`
	RecordStatus
}
