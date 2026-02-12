package models

type LabScore struct {
	AutoScore   int `json:"auto_score"`
	ManualScore int `json:"manual_score"`
}

type StudentRow struct {
	Username    string              `json:"username"`
	DisplayName string              `json:"display_name"`
	LabScores   map[string]LabScore `json:"lab_scores"`
}

type LabCol struct {
	LabID          string `json:"lab_id"`
	LabName        string `json:"lab_name"`
	MaxAutoScore   int    `json:"max_score"`
	MaxManualScore int    `json:"max_manual_score"`
}

type Gradebook struct {
	StudentRow []StudentRow `json:"student_row"`
	LabCol     []LabCol     `json:"lab_col"`
}
