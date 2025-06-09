package requests

type Course struct {
	Name     string   `json:"name" validate:"required"`
	Creators []string `json:"creators" validate:"required,uuid"`
}
