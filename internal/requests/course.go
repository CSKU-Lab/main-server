package requests

type Course struct {
	Name     string   `json:"name" validate:"required"`
	Type     string   `json:"type" validate:"required,oneof=public private"`
	Creators []string `json:"creators" validate:"required,uuid_slice"`
}
