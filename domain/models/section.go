package models

type Section struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Banner      *string             `json:"banner"`
	Semester    SectionSemester     `json:"semester"`
	Instructors []SectionInstructor `json:"instructors"`
	CourseID    string              `json:"course_id"`
}

type SectionInstructor struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
}

type SectionSemester struct {
	ID   string       `json:"id"`
	Name string       `json:"name"`
	Type SemesterType `json:"type"`
}
