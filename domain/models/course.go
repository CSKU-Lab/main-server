package models

type Course struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Creators []CourseCreator `json:"creators"`
}

type CourseCreator struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
}
