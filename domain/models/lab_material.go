package models

import (
	"time"
)

type LabMaterial struct {
	ID           string    `json:"id"`
	LabID        string    `json:"lab_id"`
	MaterialID   string    `json:"material_id"`
	Position     int       `json:"position"`
	MaterialData *Material `json:"material_data"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
