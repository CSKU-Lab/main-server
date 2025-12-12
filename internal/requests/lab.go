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

type (
	UpdateLab interface{}
	DeleteLab interface{}
)

type (
	AssignLabMaterial interface{}
	AssignLabSection  interface{}
	AssignDefaultLab  interface{}
)
