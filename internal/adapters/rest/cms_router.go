package rest

import (
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/gofiber/fiber/v2"
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
}

func NewCMSRouter(r *CMSRouter) {
	cmsRouter := r.Router.Group("/cms")

	routes.NewCmsSectionRoutes(cmsRouter, r.SectionService, r.SemesterService, r.LabSectionService)
	routes.NewAdminSemesterRoutes(cmsRouter, r.SemesterService, r.SectionService, r.CourseService)
	routes.NewCMSMaterialRoutes(cmsRouter, r.MaterialService)
	routes.NewCMSAffectedEntitiesRoutes(cmsRouter, r.AffectedEntitiesService)
	routes.NewCMSUserExistancesRoutes(cmsRouter, r.UserService)
	routes.NewCMSCourseRoutes(cmsRouter, r.CourseService, r.SectionService, r.SemesterService, r.DefaultLabService, r.LabService)
	routes.NewCMSLabRoutes(cmsRouter, r.LabService, r.LabSectionService, r.LabMaterialService)
}
