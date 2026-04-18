package models

type SidebarMaterial struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type SidebarLab struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	SubItems []*SidebarMaterial `json:"sub_items"`
}

type SidebarSection struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	CourseName string        `json:"course_name"`
	SubItems   []*SidebarLab `json:"sub_items"`
}
