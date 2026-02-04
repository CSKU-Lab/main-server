package cserrors

import "errors"

type CSError error

var (
	GroupNotFound         CSError = errors.New("GROUP_NOT_FOUND")
	SemesterNotFound      CSError = errors.New("SEMESTER_NOT_FOUND")
	SemesterAlreadyExists CSError = errors.New("SEMESTER_ALREADY_EXISTS")
	CourseAlreadyExists   CSError = errors.New("COURSE_ALREADY_EXISTS")
	UniqueViolation       CSError = errors.New("UNIQUE_VIOLATION")
	Forbidden             CSError = errors.New("FORBIDDEN")
)
