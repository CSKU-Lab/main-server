package models

type Course struct {
	ID       string          `json:"id" db:"id"`
	Name     string          `json:"name" db:"name" sort_field:"name"`
	Type     string          `json:"type" db:"type" sort_field:"type"`
	Creators []CourseCreator `json:"creators" db:"creators"`
}

type CourseCreator struct {
	ID           string  `json:"id" db:"id"`
	Username     string  `json:"username" db:"username"`
	DisplayName  string  `json:"display_name" db:"display_name"`
	ProfileImage *string `json:"profile_image" db:"profile_image"`
}
