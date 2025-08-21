package models

import "errors"

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

func (u *UserType) String() string {
	return string(*u)
}

var (
	UserTypeOauth      UserType = "oauth"
	UserTypeCredential UserType = "credential"
)
