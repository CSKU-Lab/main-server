package models

type UserGroup struct {
	ID   string `json:"id" sort_field:"id"`
	Name string `json:"name" sort_field:"name"`
}
