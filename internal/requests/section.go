package requests

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateSection struct {
	Name        string   `json:"name"`
	Instructors []string `json:"instructors"`
	Banner      *File    `json:"banner"`
	SemesterID  string   `json:"semester_id"`
	CourseID    string   `json:"course_id"`
	Students    []string `json:"students"`
}

func (req *CreateSection) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.Instructors, validation.Required, validation.Length(1, 0)),
		validation.Field(&req.Banner, validation.NilOrNotEmpty),
		validation.Field(&req.SemesterID, validation.Required),
		validation.Field(&req.CourseID, validation.Required),
		validation.Field(&req.Students, validation.NilOrNotEmpty),
	)
}

type UpdateSection struct {
	Name        string   `json:"name"`
	Instructors []string `json:"instructors"`
	SemesterID  string   `json:"semester_id"`
	Banner      *File    `json:"banner"`
	Students    []string `json:"students"`
}

func (req *UpdateSection) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Name, validation.NilOrNotEmpty),
		validation.Field(&req.Instructors, validation.NilOrNotEmpty, validation.Length(1, 0)),
		validation.Field(&req.SemesterID, validation.NilOrNotEmpty),
		validation.Field(&req.Banner, validation.NilOrNotEmpty),
		validation.Field(&req.Students, validation.NilOrNotEmpty, validation.Length(1, 0)),
	)
}

type GetSection struct {
	SectionID string `json:"section_id"`
}

func (req *GetSection) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.SectionID, validation.Required, is.UUID),
	)
}
