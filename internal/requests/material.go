package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateMaterial struct {
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Type        string   `json:"type"`
	Visibility  string   `json:"visibility"`
	ManualScore int      `json:"manual_score"`
}

func (c *CreateMaterial) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Tags),
		validation.Field(&c.Type, validation.Required, validation.In("document", "code", "typing")),
		validation.Field(&c.Visibility, validation.Required, validation.In("public", "private")),
		validation.Field(&c.ManualScore, validation.Min(0)),
	)
}

type ForkMaterial struct {
	SourceMaterialID string `json:"source_material_id"`
}

func (f *ForkMaterial) Validate() error {
	return validation.ValidateStruct(f,
		validation.Field(&f.SourceMaterialID, validation.Required),
	)
}

type BaseUpdateMaterial struct {
	Name        string    `json:"name"`
	Tags        *[]string `json:"tags"`
	Visibility  string    `json:"visibility"`
	AutoScore   *int      `json:"auto_score"`
	ManualScore *int      `json:"manual_score"`
}

func (u *BaseUpdateMaterial) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Name, validation.Skip.When(u.Name == "")),
		validation.Field(&u.Tags, validation.Skip.When(u.Tags == nil)),
		validation.Field(&u.Visibility, validation.In("public", "private"), validation.Skip.When(u.Visibility == "")),
	)
}
