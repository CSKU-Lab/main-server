package models

type Course struct {
	ID       string   `json:"id" db:"id"`
	Name     string   `json:"name" db:"name"`
	Creators []string `json:"creators" db:"creators"`
	RecordStatus
}
