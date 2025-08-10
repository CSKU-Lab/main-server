package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id" sort_field:"id"`
	Username     string    `json:"username" sort_field:"username"`
	Type         string    `json:"type" sort_field:"type"`
	Email        *string   `json:"email" sort_field:"email"`
	DisplayName  string    `json:"display_name" sort_field:"display_name"`
	ProfileImage *string   `json:"profile_image"`
	Roles        []string  `json:"roles" sort_field:"roles"`
	Group        *string   `json:"group" sort_field:"group"`
	CreatedAt    time.Time `json:"created_at" sort_field:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" sort_field:"updated_at"`
}

type UserType string

func (u *UserType) String() string {
	return string(*u)
}

var (
	UserTypeOauth      UserType = "oauth"
	UserTypeCredential UserType = "credential"
)
