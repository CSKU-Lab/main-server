package models

import (
	"time"
)

type LabMaterial struct {
	ID         string    `json:"id"`
	LabID      string    `json:"lab_id"`
	MaterialID string    `json:"material_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
