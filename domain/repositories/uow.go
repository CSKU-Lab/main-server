package repositories

import "context"

type UoWInstance interface {
	User() User
	UserPassword() UserPassword
	UserGroup() UserGroup

	Section() SectionRepository
	SectionInstructor() SectionInstructorRepository
	SectionStudent() SectionStudentRepository

	Course() CourseRepository
	CourseCreator() CourseCreatorRepository

	Material() MaterialRepository
	MaterialTag() WriteMaterialTagRepository
}

type UoWRepository interface {
	Execute(ctx context.Context, cb func(u UoWInstance) error) error
}
