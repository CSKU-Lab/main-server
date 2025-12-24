package models

import "time"

type SectionLog struct {
	ID        string         `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Action    string         `json:"action"`
	User      SectionLogUser `json:"user"`
	IPAddress string         `json:"ip_address"`
}

type SectionLogUser struct {
	Username     string   `json:"username"`
	DisplayName  string   `json:"display_name"`
	ProfileImage *string  `json:"profile_image"`
	Roles        []string `json:"roles"`
}
