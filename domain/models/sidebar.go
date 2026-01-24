package models

type Sidebar struct {
	Name     string     `json:"name"`
	ID       string     `json:"id"`
	Status   string     `json:"status"`
	SubItems []*Sidebar `json:"sub_items"`
}
