package repositories

import "context"

type SectionUoWInstance interface {
	Section() SectionRepository
	SectionInstructor() SectionInstructorRepository
	SectionStudent() SectionStudentRepository
}

type SectionUoWRepository interface {
	Execute(ctx context.Context, cb func(s SectionUoWInstance) error) error
}
