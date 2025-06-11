package models

type Course struct {
	ID       string          `json:"id" db:"id"`
	Name     string          `json:"name" db:"name"`
	Creators []CourseCreator `json:"creators" db:"creators"`
	RecordStatus
}

type CourseCreator struct {
	ID           string  `json:"id" db:"id"`
	DisplayName  string  `json:"display_name" db:"display_name"`
	ProfileImage *string `json:"profile_image" db:"profile_image"`
}
