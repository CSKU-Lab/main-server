package models

type EntityDetail struct {
	Name     string           `json:"name"`
	Children []AffectedEntity `json:"children"`
}

type AffectedEntity struct {
	Type string         `json:"type"`
	Data []EntityDetail `json:"data"`
}
