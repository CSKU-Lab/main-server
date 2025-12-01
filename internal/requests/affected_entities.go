package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type GetAffectedEntities struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (c *GetAffectedEntities) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ID, validation.Required),
		validation.Field(&c.Type, validation.Required),
	)
}
