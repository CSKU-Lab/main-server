package models

import validation "github.com/go-ozzo/ozzo-validation/v4"

type SemesterType string

const (
	FirstSemesterType  SemesterType = "first"
	SecondSemesterType SemesterType = "second"
	SummerSemesterType SemesterType = "summer"
)

func (s *SemesterType) Validate() error {
	return validation.Validate(s, validation.Required, validation.In("first", "second", "summer").Error("must be one of 'first', 'second', 'summer'"))
}
