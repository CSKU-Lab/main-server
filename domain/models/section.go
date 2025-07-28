package models

type Section struct {
	ID        string
	Name      string
	Image     *string
	StartedAt string
	EndedAt   string
	RecordStatus
}
