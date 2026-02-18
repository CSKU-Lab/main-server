package models

import (
	"time"
)

type LabSection struct {
	ID        string     `json:"id"`
	LabID     string     `json:"lab_id"`
	SectionID string     `json:"section_id"`
	Position  int        `json:"position"`
	Status    string     `json:"status"`
	OpenedAt  *time.Time `json:"opened_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
