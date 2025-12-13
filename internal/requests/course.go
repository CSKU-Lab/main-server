package requests

import (
	"errors"

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
	Name     *string   `json:"name"`
	Creators *[]string `json:"creators"`
	Type     *string   `json:"type"`
}

func (c *UpdateCourse) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Skip.When(c.Name == nil)),
		validation.Field(&c.Creators, validation.Skip.When(c.Creators == nil), validation.By(func(value any) error {
			val, ok := value.(*[]string)
			if !ok {
				return errors.New("invalid type")
			}

			if val == nil {
				return nil
			}

			if len(*val) == 0 {
				return errors.New("creators cannot be empty")
			}

			for _, v := range *val {
				if err := is.UUID.Validate(v); err != nil {
					return errors.New("each creator must be a valid UUID")
				}
			}

			return nil
		})),
		validation.Field(&c.Type, validation.Skip.When(c.Type == nil), validation.In("public", "private")),
	)
}
