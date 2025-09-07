package models

type Section struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Image     *string `json:"image"`
	StartedAt string  `json:"started_at"`
	EndedAt   string  `json:"ended_at"`
	RecordStatus
}
