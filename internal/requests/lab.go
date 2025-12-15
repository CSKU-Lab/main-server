package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateLab struct {
	DisplayName string `json:"display_name"`
	CourseID    string `json:"course_id"`
}

func (l *CreateLab) Validate() error {
	return validation.ValidateStruct(l,
		validation.Field(&l.DisplayName, validation.Required),
		validation.Field(&l.CourseID, validation.Required, is.UUID),
	)
}

type BaseUpdateLab struct {
	DisplayName string `json:"display_name"`
	CourseID    string `json:"course_id"`
}

func (l *BaseUpdateLab) Validate() error {
	return validation.ValidateStruct(l,
		validation.Field(&l.DisplayName, validation.NilOrNotEmpty),
		validation.Field(&l.CourseID, validation.Skip.When(l.CourseID == ""), is.UUID),
	)
}

type (
	SetLabSection struct {
		LabID     string `json:"lab_id"`
		SectionID string `json:"section_id"`
		Position  int    `json:"position"`
	}
	UpdateLabSection struct {
		Position int `json:"position"`
	}
)

type (
	SetLabMaterial struct{}
)

type (
	SetDefaultLab struct{}
)

func (ls *SetLabSection) Validate() error {
	return validation.ValidateStruct(ls,
		validation.Field(&ls.LabID, validation.Required, is.UUID),
		validation.Field(&ls.SectionID, validation.Required, is.UUID),
		validation.Field(
			&ls.Position,
			validation.Required,
			validation.Min(1),
		),
	)
}

func (ls *UpdateLabSection) Validate() error {
	return validation.ValidateStruct(ls,
		validation.Field(
			&ls.Position,
			validation.Required,
			validation.Min(1),
		),
	)
}
