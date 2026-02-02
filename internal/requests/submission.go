package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Submission struct {
	LabID      string  `json:"lab_id"`
	MaterialID string  `json:"material_id"`
	SectionID  *string `json:"section_id"`
	CourseID   *string `json:"course_id"`
	Payload    any     `json:"payload"`
}

func (s *Submission) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.LabID, validation.Required, is.UUID),
		validation.Field(&s.MaterialID, validation.Required, is.UUID),
		validation.Field(&s.SectionID, validation.When(s.CourseID != nil, validation.Nil).Else(validation.Required, is.UUID)),
		validation.Field(&s.CourseID, validation.When(s.SectionID != nil, validation.Nil).Else(validation.Required, is.UUID)),
		validation.Field(&s.Payload, validation.Required),
	)
}
