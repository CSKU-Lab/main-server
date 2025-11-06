package models

type Material struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Tags       []string `json:"tags"`
	Type       string   `json:"type"`
	Visibility string   `json:"visibility"`
}
