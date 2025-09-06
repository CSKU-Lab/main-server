package models

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type UserType string

func (u UserType) New(str string) (UserType, error) {
	switch str {
	case "oauth":
		return UserTypeOauth, nil
	case "credential":
		return UserTypeCredential, nil
	}

	return "", errors.New("unknown user type")
}

func (u UserType) Validate() error {
	return validation.Validate(string(u), validation.Required, validation.In("oauth", "credential").Error("must be one of 'oauth', 'credential'"))
}

var (
	UserTypeOauth      UserType = "oauth"
	UserTypeCredential UserType = "credential"
)
