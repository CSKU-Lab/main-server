package requests

import validation "github.com/go-ozzo/ozzo-validation/v4"

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Credential) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Username, validation.Required),
		validation.Field(&c.Password, validation.Required),
	)
}
