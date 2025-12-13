package models

type Course struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Visibility string          `json:"visibility"`
	Creators   []CourseCreator `json:"creators"`
}

type CourseCreator struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
}
