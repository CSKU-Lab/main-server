package requests

import validation "github.com/go-ozzo/ozzo-validation/v4"

type Submission struct {
	MaterialID string  `json:"material_id"`
	SectionID  *string `json:"section_id"`
	CourseID   *string `json:"course_id"`
	Type       string  `json:"type"`
	Payload    any     `json:"payload"`
}

func (s *Submission) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.MaterialID, validation.Required),
		validation.Field(&s.SectionID, validation.Skip.When(s.CourseID != nil), validation.Required),
		validation.Field(&s.CourseID, validation.Skip.When(s.SectionID != nil), validation.Required),
		validation.Field(&s.Type, validation.Required),
		validation.Field(&s.Payload, validation.Required),
	)
}
