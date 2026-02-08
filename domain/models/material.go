package models

import "time"

type Material struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Tags       []string         `json:"tags"`
	Type       string           `json:"type"`
	Visibility string           `json:"visibility"`
	CreatedAt  time.Time        `json:"created_at"`
	CreatedBy  *MaterialCreator `json:"created_by"`
}

type MaterialCreator struct {
	ID           string  `json:"id"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
}

type MaterialDetail struct {
	*Material
	Payload any `json:"payload"`
}
