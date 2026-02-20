package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type UpdateSubmissionManualScore struct {
	ManualScore int `json:"manual_score"`
}

func (r *UpdateSubmissionManualScore) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.ManualScore, validation.Required, validation.Min(0)),
	)
}
