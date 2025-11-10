package models

import "time"

type Material struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Tags       []string  `json:"tags"`
	Type       string    `json:"type"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
}

type MaterialDetail struct {
	*Material
	Payload any `json:"payload"`
}
