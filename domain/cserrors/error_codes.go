package cserrors

import "errors"

var (
	GroupNotFound         = errors.New("GROUP_NOT_FOUND")
	SemesterNotFound      = errors.New("SEMESTER_NOT_FOUND")
	SemesterAlreadyExists = errors.New("SEMESTER_ALREADY_EXISTS")
)
