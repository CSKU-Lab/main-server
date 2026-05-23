package models

import (
	"time"
)

type LabSection struct {
	ID         string     `json:"id"`
	LabID      string     `json:"lab_id"`
	SectionID  string     `json:"section_id"`
	Position   int        `json:"position"`
	Status     string     `json:"status"`
	OpenedAt   *time.Time `json:"opened_at"`
	ReadonlyAt *time.Time `json:"readonly_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (ls *LabSection) EffectiveStatus() string {
	if ls.OpenedAt == nil && ls.ReadonlyAt == nil {
		return ls.Status
	}
	now := time.Now()
	if ls.OpenedAt != nil && now.Before(*ls.OpenedAt) {
		return "hidden"
	}
	if ls.ReadonlyAt != nil && now.After(*ls.ReadonlyAt) {
		return "readonly"
	}
	return "open"
}
