package models

import "time"

type InputSubmission struct {
	ID                 string    `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	NodeID             string    `json:"node_id" db:"node_id"`
	DocumentMaterialID string    `json:"document_material_id" db:"document_material_id"`
	LabID              string    `json:"lab_id" db:"lab_id"`
	SectionID          *string   `json:"section_id" db:"section_id"`
	Value              string    `json:"value" db:"value"`
	Passed             bool      `json:"passed" db:"passed"`
	Score              int       `json:"score" db:"score"`
	Graded             bool      `json:"graded" db:"graded"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}
