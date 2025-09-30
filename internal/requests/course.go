package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateCourse struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Creators []string `json:"creators"`
}

func (c *CreateCourse) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Type, validation.Required, validation.In("public", "private")),
		validation.Field(&c.Creators, validation.Required, validation.Length(1, 0), validation.Each(is.UUID)),
	)
}

type UpdateCourse struct {
	Name     string   `json:"name"`
	Creators []string `json:"creators"`
}

func (c *UpdateCourse) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Skip.When(c.Name == "")),
		validation.Field(&c.Creators, validation.Skip.When(len(c.Creators) == 0), validation.Length(1, 0), validation.Each(is.UUID)),
	)
}
