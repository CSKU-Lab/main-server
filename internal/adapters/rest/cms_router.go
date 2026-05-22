package rest

import (
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/CSKU-Lab/queue"
	"github.com/gofiber/fiber/v3"
)

type CMSRouter struct {
	Router                  fiber.Router
	UserService             services.UserService
	SemesterService         services.SemesterService
	CourseService           services.CourseService
	SectionService          services.SectionService
	MaterialService         services.MaterialService
	AffectedEntitiesService services.AffectedEntitiesService
	LabService              services.LabService
	LabSectionService       services.LabSectionService
	LabMaterialService      services.LabMaterialService
	DefaultLabService       services.DefaultLabService
	SectionLogService       services.SectionLogService
	MaterialAssetService    services.MaterialAssetService
	ConfigGRPCClient        configPB.ConfigServiceClient
	SubmissionService       services.SubmissionService
	GradebookExportService  services.GradebookExportService
	SearchService           services.SearchService
	Queue                   queue.Queue
	PermissionService       permission.Service
	TagService              services.TagService
}

func NewCMSRouter(r *CMSRouter) {
	cmsRouter := r.Router.Group("/cms")

	routes.NewCmsSectionRoutes(cmsRouter, r.SectionService, r.SemesterService, r.LabSectionService, r.SectionLogService, r.LabService, r.SubmissionService, r.GradebookExportService, r.PermissionService)
	routes.NewAdminSemesterRoutes(cmsRouter, r.SemesterService, r.SectionService, r.CourseService)
	routes.NewCMSMaterialRoutes(cmsRouter, r.MaterialService, r.MaterialAssetService, r.SubmissionService, r.PermissionService)
	routes.NewCMSAffectedEntitiesRoutes(cmsRouter, r.AffectedEntitiesService)
	routes.NewCMSUserExistancesRoutes(cmsRouter, r.UserService)
	routes.NewCMSCourseRoutes(cmsRouter, r.CourseService, r.SectionService, r.SemesterService, r.DefaultLabService, r.LabService, r.PermissionService)
	routes.NewCMSLabRoutes(cmsRouter, r.LabService, r.LabSectionService, r.LabMaterialService)
	routes.NewCMSConfigRoutes(cmsRouter, r.ConfigGRPCClient, r.Queue)
	routes.NewCMSSubmissionRoutes(cmsRouter, r.SubmissionService, r.PermissionService)
	routes.NewCMSUserRoute(cmsRouter, r.UserService)
	routes.NewCMSTagRoutes(cmsRouter, r.TagService)
	routes.NewCMSSearchRoutes(cmsRouter, r.SearchService)
}
