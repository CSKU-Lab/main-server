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
	CourseEnrollment() CourseEnrollmentRepository

	Material() MaterialRepository
	MaterialTag() WriteMaterialTagRepository

	Lab() LabRepository
	LabSection() LabSectionRepository
	LabMaterial() LabMaterialRepository

	DefaultLab() DefaultLabRepository

	Submission() SubmissionRepository
	CodeSubmission() CodeSubmissionRepository
	CodeSubmissionOutbox() CodeSubmissionOutboxRepository

	TypingSubmission() TypingSubmissionRepository
}

type UoWRepository interface {
	Execute(ctx context.Context, cb func(u UoWInstance) error) error
}
