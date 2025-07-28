package requests

type Section struct {
	Name      string
	Image     *File
	StartedAt string
	EndedAt   string
}

type CreateSection struct {
	Section
	CourseID   string
	SemesterID string
}
