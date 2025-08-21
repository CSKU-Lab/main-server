package models

type UserType string

func (u *UserType) New(str string) UserType {
	switch str {
	case "oauth":
		return UserTypeOauth
	case "credential":
		return UserTypeCredential
	}

	panic("wrong user type")
}

func (u *UserType) String() string {
	return string(*u)
}

var (
	UserTypeOauth      UserType = "oauth"
	UserTypeCredential UserType = "credential"
)
