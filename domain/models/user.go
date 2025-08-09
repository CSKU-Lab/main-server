package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Type         string    `json:"type"`
	Email        *string   `json:"email"`
	DisplayName  string    `json:"display_name"`
	ProfileImage *string   `json:"profile_image"`
	Roles        []string  `json:"roles"`
	Group        *string   `json:"group"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserType string

func (u *UserType) String() string {
	return string(*u)
}

var (
	UserTypeOauth      UserType = "oauth"
	UserTypeCredential UserType = "credential"
)
