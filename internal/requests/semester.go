package requests

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Semester struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	StartedDate time.Time `json:"started_date"`
}

func (s *Semester) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required),
		validation.Field(&s.Type, validation.Required, validation.In("first", "second", "summer").Error("must be one of 'first', 'second', 'summer'")),
		validation.Field(&s.StartedDate, validation.Required),
	)
}

type UpdateSemester struct {
	Name        *string    `json:"name"`
	Type        *string    `json:"type"`
	StartedDate *time.Time `json:"started_date"`
}

func (s *UpdateSemester) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.NilOrNotEmpty),
		validation.Field(&s.Type, validation.NilOrNotEmpty, validation.In("first", "second", "summer").Error("must be one of 'first', 'second', 'summer'")),
		validation.Field(&s.StartedDate, validation.NilOrNotEmpty),
	)
}
