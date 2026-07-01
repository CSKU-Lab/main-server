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

func NewUserAuthProviderRepository(db *sqlx.DB) repositories.UserAuthProviderRepository {
	return sqlxAdapter.NewUserAuthProviderRepository(db)
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

func NewInputSubmissionRepository(db *sqlx.DB) repositories.InputSubmissionRepository {
	return sqlxAdapter.NewInputSubmissionRepository(db)
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

func NewDocumentMaterialRepository(db *sqlx.DB) repositories.DocumentMaterialRepository {
	return sqlxAdapter.NewDocumentMaterialRepository(db)
}

func NewSearchRepository(db *sqlx.DB) repositories.SearchRepository {
	return sqlxAdapter.NewSearchRepository(db)
}

func NewSystemSettingsRepository(db *sqlx.DB) repositories.SystemSettingsRepository {
	return sqlxAdapter.NewSystemSettingsRepository(db)
}

func NewAnalyticsRepository(db *sqlx.DB) repositories.AnalyticsRepository {
	return sqlxAdapter.NewAnalyticsRepository(db)
}

func NewAuthLogRepository(db *sqlx.DB) repositories.AuthLogRepository {
	return sqlxAdapter.NewAuthLogRepository(db)
}

func NewUserActivityRepository(db *sqlx.DB) repositories.UserActivityRepository {
	return sqlxAdapter.NewUserActivityRepository(db)
}

var RepositorySet = wire.NewSet(
	sqlxAdapter.NewUoWRepository,
	NewUserRepository,
	NewUserPasswordRepository,
	NewUserGroupRepository,
	NewUserAuthProviderRepository,
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
	NewInputSubmissionRepository,
	NewSqlxLabRepository,
	NewSqlxLabSectionRepository,
	NewSqlxLabMaterialRepository,
	NewSqlxDefaultLabRepository,
	NewSubmissionRepository,
	NewCodeSubmissionRepository,
	NewDocumentMaterialRepository,
	NewSearchRepository,
	NewSystemSettingsRepository,
	NewAnalyticsRepository,
	NewAuthLogRepository,
	NewUserActivityRepository,
)
