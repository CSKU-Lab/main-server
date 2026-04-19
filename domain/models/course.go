package models

type Course struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   *string         `json:"description"`
	Banner        *string         `json:"banner"`
	Visibility    string          `json:"visibility"`
	TotalStudents int             `json:"total_students"`
	Creators      []CourseCreator `json:"creators"`
}

type CourseCreator struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
}
