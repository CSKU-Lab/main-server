package models

import "time"

type Material struct {
	ID                   string           `json:"id"`
	CourseID             string           `json:"course_id"`
	ForkedFromMaterialID *string          `json:"forked_from_material_id"`
	Name                 string           `json:"name"`
	Tags                 []string         `json:"tags"`
	Type                 string           `json:"type"`
	Visibility           string           `json:"visibility"`
	AutoScore            int              `json:"auto_score"`
	ManualScore          int              `json:"manual_score"`
	CreatedAt            time.Time        `json:"created_at"`
	CreatedBy            *MaterialCreator `json:"created_by"`
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
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	LabStatus   string `json:"lab_status"`
	ManualScore int    `json:"manual_score"`
	Payload     any    `json:"payload"`
}
