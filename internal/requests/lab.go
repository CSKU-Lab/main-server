package requests

import (
	"time"

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
	IsDefault   *bool  `json:"is_default"`
}

func (l *BaseUpdateLab) Validate() error {
	return validation.ValidateStruct(l,
		validation.Field(&l.DisplayName, validation.Skip.When(l.DisplayName == ""), validation.Length(1, 0)),
		validation.Field(&l.CourseID, validation.Skip.When(l.CourseID == ""), is.UUID),
		validation.Field(&l.IsDefault, validation.Skip.When(l.IsDefault == nil), validation.In(true, false)),
	)
}

type (
	SetLabSection struct {
		LabIDs []string `json:"lab_ids"`
	}
	UpdateLabSection struct {
		Position int    `json:"position"`
		LabID    string `json:"lab_id"`
	}
	UpdateLabSectionStatus struct {
		Status   *string    `json:"status"`
		OpenedAt *time.Time `json:"opened_at"`
		ClosedAt *time.Time `json:"closed_at"`
	}
	DeleteLabSection struct {
		LabIDs []string `json:"lab_ids"`
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
		Position *int   `json:"position"`
	}
	UpdateDefaultLab struct {
		LabID    string `json:"lab_id"`
		Position int    `json:"position"`
	}
	DeleteDefaultLab struct {
		LabID string `json:"lab_id"`
	}
)

func (ls *SetLabSection) Validate() error {
	return validation.ValidateStruct(ls,
		validation.Field(&ls.LabIDs, validation.Required, validation.Length(1, 0), validation.Each(is.UUID)),
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

func (ls *UpdateLabSectionStatus) Validate() error {
	return validation.ValidateStruct(ls,
		validation.Field(
			&ls.Status,
			validation.Skip.When(ls.Status == nil),
			validation.In("hidden", "open", "readonly", "disabled"),
		),
		validation.Field(&ls.OpenedAt, validation.Skip.When(ls.OpenedAt == nil)),
		validation.Field(&ls.ClosedAt, validation.Skip.When(ls.ClosedAt == nil)),
	)
}

func (ls *DeleteLabSection) Validate() error {
	return validation.ValidateStruct(ls,
		validation.Field(&ls.LabIDs, validation.Required, validation.Length(1, 0), validation.Each(is.UUID)),
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
			validation.NilOrNotEmpty,
			validation.Min(1),
		),
	)
}

func (dl *UpdateDefaultLab) Validate() error {
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
