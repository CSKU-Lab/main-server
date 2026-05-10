package providers

import (
	"github.com/CSKU-Lab/main-server/domain/repositories"
	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"

	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
)

func NewUserRepository(db *sqlx.DB) repositories.User {
	return sqlxAdapter.NewUserRepository(db)
}

func NewUserPasswordRepository(db *sqlx.DB) repositories.UserPassword {
	return sqlxAdapter.NewUserPasswordRepository(db)
}

func NewUserGroupRepository(db *sqlx.DB) repositories.UserGroup {
	return sqlxAdapter.NewUserGroupRepository(db)
}

func NewCourseRepository(db *sqlx.DB) repositories.CourseRepository {
	return sqlxAdapter.NewCourseRepository(db)
}

func NewCourseCreatorRepository(db *sqlx.DB) repositories.CourseCreatorRepository {
	return sqlxAdapter.NewCourseCreatorRepository(db)
}

func NewCourseEnrollmentRepository(db *sqlx.DB) repositories.CourseEnrollmentRepository {
	return sqlxAdapter.NewCourseEnrollmentRepository(db)
}

func NewSectionRepository(db *sqlx.DB) repositories.SectionRepository {
	return sqlxAdapter.NewSectionRepository(db)
}

func NewSectionInstructorRepository(db *sqlx.DB) repositories.SectionInstructorRepository {
	return sqlxAdapter.NewSectionInstructorRepository(db)
}

func NewSectionStudentRepository(db *sqlx.DB) repositories.SectionStudentRepository {
	return sqlxAdapter.NewSectionStudentRepository(db)
}

func NewSectionLogRepository(db *sqlx.DB) repositories.SectionLogRepository {
	return sqlxAdapter.NewSectionLogRepository(db)
}

func NewMaterialRepository(db *sqlx.DB) repositories.MaterialRepository {
	return sqlxAdapter.NewMaterialRepository(db)
}

func NewTypingSubmissionRepository(db *sqlx.DB) repositories.TypingSubmissionRepository {
	return sqlxAdapter.NewTypingSubmissionRepository(db)
}

func NewSqlxLabRepository(db *sqlx.DB) repositories.LabRepository {
	return sqlxAdapter.NewSqlxLabRepository(db)
}

func NewSqlxLabSectionRepository(db *sqlx.DB) repositories.LabSectionRepository {
	return sqlxAdapter.NewSqlxLabSectionRepository(db)
}

func NewSqlxLabMaterialRepository(db *sqlx.DB) repositories.LabMaterialRepository {
	return sqlxAdapter.NewSqlxLabMaterialRepository(db)
}

func NewSqlxDefaultLabRepository(db *sqlx.DB) repositories.DefaultLabRepository {
	return sqlxAdapter.NewSqlxDefaultLabRepository(db)
}

func NewSubmissionRepository(db *sqlx.DB) repositories.SubmissionRepository {
	return sqlxAdapter.NewSubmissionRepository(db)
}

func NewCodeSubmissionRepository(db *sqlx.DB) repositories.CodeSubmissionRepository {
	return sqlxAdapter.NewCodeSubmission(db)
}

var RepositorySet = wire.NewSet(
	sqlxAdapter.NewUoWRepository,
	NewUserRepository,
	NewUserPasswordRepository,
	NewUserGroupRepository,
	sqlxAdapter.NewSQLxRefreshTokenRepository,
	NewCourseRepository,
	NewCourseCreatorRepository,
	NewCourseEnrollmentRepository,
	NewSectionRepository,
	NewSectionInstructorRepository,
	NewSectionStudentRepository,
	NewSectionLogRepository,
	sqlxAdapter.NewSqlxSemesterRepository,
	NewMaterialRepository,
	sqlxAdapter.NewReadMaterialTagRepository,
	sqlxAdapter.NewTagRepository,
	sqlxAdapter.NewCodeMaterialRepository,
	sqlxAdapter.NewTypingMaterialRepository,
	NewTypingSubmissionRepository,
	NewSqlxLabRepository,
	NewSqlxLabSectionRepository,
	NewSqlxLabMaterialRepository,
	NewSqlxDefaultLabRepository,
	NewSubmissionRepository,
	NewCodeSubmissionRepository,
)
