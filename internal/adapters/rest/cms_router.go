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
}

func NewCMSRouter(r *CMSRouter) {
	cmsRouter := r.Router.Group("/cms")

	routes.NewCmsSectionRoutes(cmsRouter, r.SectionService, r.SemesterService)
	routes.NewAdminSemesterRoutes(cmsRouter, r.SemesterService, r.SectionService, r.CourseService)
	routes.NewCMSMaterialRoutes(cmsRouter, r.MaterialService)
	routes.NewCMSAffectedEntitiesRoutes(cmsRouter, r.AffectedEntitiesService)
	routes.NewCMSUserExistancesRoutes(cmsRouter, r.UserService)
	routes.NewCMSCourseRoutes(cmsRouter, r.SectionService, r.SemesterService)
}
