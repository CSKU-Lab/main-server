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
		LabID    string `json:"lab_id"`
		Position int    `json:"position"`
	}
	UpdateLabSection struct {
		Position int    `json:"position"`
		LabID    string `json:"lab_id"`
	}
	DeleteLabSection struct {
		LabID string `json:"lab_id"`
	}
)

type (
	SetLabMaterial struct {
		MaterialID string `json:"material_id"`
	}
	DeleteLabMaterial struct {
		MaterialID string `json:"material_id"`
	}
)

type (
	SetDefaultLab struct {
		LabID    string `json:"lab_id"`
		Position int    `json:"position"`
	}
	DeleteDefaultLab struct {
		LabID string `json:"lab_id"`
	}
)

func (ls *SetLabSection) Validate() error {
	return validation.ValidateStruct(ls,
		validation.Field(&ls.LabID, validation.Required, is.UUID),
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
		validation.Field(&ls.LabID, validation.Required, is.UUID),
	)
}

func (ls *DeleteLabSection) Validate() error {
	return validation.ValidateStruct(ls,
		validation.Field(&ls.LabID, validation.Required, is.UUID),
	)
}

func (lm *SetLabMaterial) Validate() error {
	return validation.ValidateStruct(lm,
		validation.Field(&lm.MaterialID, validation.Required, is.UUID),
	)
}

func (lm *DeleteLabMaterial) Validate() error {
	return validation.ValidateStruct(lm,
		validation.Field(&lm.MaterialID, validation.Required, is.UUID),
	)
}

func (dl *SetDefaultLab) Validate() error {
	return validation.ValidateStruct(dl,
		validation.Field(&dl.LabID, validation.Required, is.UUID),
		validation.Field(
			&dl.Position,
			validation.Required,
			validation.Min(1),
		),
	)
}

func (dl *DeleteDefaultLab) Validate() error {
	return validation.ValidateStruct(dl,
		validation.Field(&dl.LabID, validation.Required, is.UUID),
	)
}
