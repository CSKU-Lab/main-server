package repositories

import "context"

type UoWInstance interface {
	User() User
	UserPassword() UserPassword
	UserGroup() UserGroup

	Section() SectionRepository
	SectionInstructor() SectionInstructorRepository
	SectionStudent() SectionStudentRepository
	SectionLog() SectionLogRepository

	Course() CourseRepository
	CourseCreator() CourseCreatorRepository

	Material() MaterialRepository
	MaterialTag() WriteMaterialTagRepository

	Lab() LabRepository
	LabSection() LabSectionRepository

	DefaultLab() DefaultLabRepository
}

type UoWRepository interface {
	Execute(ctx context.Context, cb func(u UoWInstance) error) error
}
