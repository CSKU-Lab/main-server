package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type SectionStudents struct {
	StudentUsernames []string `json:"student_usernames"`
}

func (s *SectionStudents) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.StudentUsernames, validation.Required, validation.Length(1, 0)),
	)
}

type RemoveStudent struct {
	StudentIDs []string `json:"student_ids"`
}

func (r *RemoveStudent) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.StudentIDs, validation.Required, validation.Length(1, 0), validation.Each(is.UUID)),
	)
}
