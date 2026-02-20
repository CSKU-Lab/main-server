package models

import "time"

type Material struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Tags           []string         `json:"tags"`
	Type           string           `json:"type"`
	Visibility     string           `json:"visibility"`
	MaxAutoScore   int              `json:"max_auto_score"`
	MaxManualScore int              `json:"max_manual_score"`
	CreatedAt      time.Time        `json:"created_at"`
	CreatedBy      *MaterialCreator `json:"created_by"`
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

type MaterialWithSubmissionStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Payload any    `json:"payload"`
}
