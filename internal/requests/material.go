package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateMaterial struct {
	Name       string   `json:"name"`
	Tags       []string `json:"tags"`
	Type       string   `json:"type"`
	Visibility string   `json:"visibility"`
}

func (c *CreateMaterial) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Tags),
		validation.Field(&c.Type, validation.Required, validation.In("document", "code", "type")),
		validation.Field(&c.Visibility, validation.Required, validation.In("public", "private")),
	)
}

type UpdateMaterial struct {
	Name       string   `json:"name"`
	Tags       []string `json:"tags"`
	Type       string   `json:"type"`
	Visibility string   `json:"visibility"`
}

func (u *UpdateMaterial) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Name, validation.Skip.When(u.Name == "")),
		validation.Field(&u.Tags, validation.Skip.When(len(u.Tags) == 0)),
		validation.Field(&u.Type, validation.In("document", "code", "type"), validation.Skip.When(u.Type == "")),
		validation.Field(&u.Visibility, validation.In("public", "private"), validation.Skip.When(u.Visibility == "")),
	)
}
