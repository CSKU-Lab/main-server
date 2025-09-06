package requests

import (
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateSemester struct {
	Name        string          `json:"name"`
	Type        models.Semester `json:"type"`
	StartedDate time.Time       `json:"started_date"`
}

func (s *CreateSemester) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required),
		validation.Field(&s.Type),
		validation.Field(&s.StartedDate, validation.Required),
	)
}

type UpdateSemester struct {
	Name        string              `json:"name"`
	Type        models.SemesterType `json:"type"`
	StartedDate time.Time           `json:"started_date"`
}

func (s *UpdateSemester) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.NilOrNotEmpty),
		validation.Field(&s.Type),
		validation.Field(&s.StartedDate, validation.NilOrNotEmpty),
	)
}
