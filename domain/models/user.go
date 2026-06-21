package models

import (
	"time"
)

type User struct {
	ID            string         `json:"id"`
	Username      string         `json:"username"`
	Type          string         `json:"type"`
	Email         *string        `json:"email"`
	DisplayName   string         `json:"display_name"`
	ProfileImage  *string        `json:"profile_image"`
	Roles         []Role         `json:"roles"`
	Group         *UserGroup     `json:"group"`
	AuthProviders []AuthProvider `json:"auth_providers"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type CMSUser struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
}
