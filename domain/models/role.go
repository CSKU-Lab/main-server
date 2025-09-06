package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Role string

const (
	ADMIN      Role = "admin"
	STUDENT    Role = "student"
	INSTRUCTOR Role = "instructor"
)

func (r Role) Validate() error {
	return validation.Validate(string(r), validation.Required, validation.In("admin", "instructor", "student"))
}
