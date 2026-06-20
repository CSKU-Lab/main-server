package providers

import (
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"

	"github.com/google/wire"
)

func NewSubmissionServiceArgs(
	submissionRepo repositories.SubmissionRepository,
	materialRepo repositories.MaterialRepository,
	uowRepo repositories.UoWRepository,
	submissionRegistry registries.SubmissionRegistry,
	sectionStudentRepo repositories.SectionStudentRepository,
	userRepo repositories.User,
	materialRegistry registries.Material,
	sectionRepo repositories.SectionRepository,
	labSectionRepo repositories.LabSectionRepository,
	labMaterialRepo repositories.LabMaterialRepository,
	codeSubmissionRepo repositories.CodeSubmissionRepository,
	codeMatRepo repositories.CodeMaterialRepository,
	pubSub pubsub.PubSub,
	systemSettingsService services.SystemSettingsService,
) *services.SubmissionServiceArgs {
	return &services.SubmissionServiceArgs{
		SubmissionRepository:     submissionRepo,
		MaterialRepository:       materialRepo,
		UowRepository:            uowRepo,
		SubmissionRegistry:       submissionRegistry,
		SectionStudentRepository: sectionStudentRepo,
		UserRepository:           userRepo,
		MaterialRegistry:         materialRegistry,
		SectionRepository:        sectionRepo,
		LabSectionRepository:     labSectionRepo,
		LabMaterialRepository:    labMaterialRepo,
		CodeSubmissionRepository: codeSubmissionRepo,
		CodeMaterialRepository:   codeMatRepo,
		PubSub:                   pubSub,
		SystemSettingsService:    systemSettingsService,
	}
}

var ServiceSet = wire.NewSet(
	services.NewSystemSettingsService,
	services.NewSearchService,
	services.NewUserService,
	services.NewUserGroupService,
	services.NewRefreshTokenService,
	services.NewCourseService,
	services.NewCourseEnrollmentService,
	services.NewSemesterService,
	services.NewSectionLogService,
	services.NewSectionService,
	services.NewSectionStudentService,
	services.NewTagService,
	services.NewMaterialAssetService,
	services.NewLabService,
	services.NewLabSectionService,
	services.NewLabMaterialService,
	services.NewDefaultLabService,
	services.NewAffectedEntitiesService,
	NewSubmissionServiceArgs,
	services.NewSubmissionService,
	services.NewSidebarService,
	services.NewGradebookExportService,
	services.NewMaterialService,
	permission.NewService,
)
