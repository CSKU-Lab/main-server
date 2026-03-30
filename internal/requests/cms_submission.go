package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type UpdateSubmissionManualScore struct {
	ManualScore int `json:"manual_score"`
}

func (r *UpdateSubmissionManualScore) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.ManualScore, validation.Min(0)),
	)
}

// ValidateWithMax validates the manual score against a maximum value
func (r *UpdateSubmissionManualScore) ValidateWithMax(maxScore int) error {
	return validation.ValidateStruct(r,
		validation.Field(&r.ManualScore,
			validation.Min(0),
			validation.Max(maxScore),
		),
	)
}
