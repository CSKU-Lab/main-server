package rest

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/gofiber/fiber/v3"
)

type CoreRouter struct {
	Router                fiber.Router
	SectionService        services.SectionService
	LabSectionService     services.LabSectionService
	LabService            services.LabService
	SectionStudentService services.SectionStudentService
	LabMaterialService    services.LabMaterialService
	CourseService         services.CourseService
	SidebarService        services.SidebarService
	MaterialService       services.MaterialService
	SubmissionService     services.SubmissionService
	PubSub                pubsub.PubSub
	PermissionService     permission.Service
}

func NewCoreRouter(r *CoreRouter) {
	coreRouter := r.Router.Group("", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
		models.STUDENT,
	}))

	routes.NewCoreSectionRoute(
		coreRouter,
		r.SectionService,
		r.LabSectionService,
		r.LabService,
		r.SectionStudentService,
		r.LabMaterialService,
		r.CourseService,
	)
	routes.NewCoreLabRoute(
		coreRouter,
		r.SectionService,
		r.LabSectionService,
		r.LabService,
		r.SectionStudentService,
		r.LabMaterialService,
	)
	routes.NewCoreSidebarRoute(
		coreRouter,
		r.SidebarService,
	)

	routes.NewCoreSubmissionRoutes(
		coreRouter,
		r.SubmissionService,
		r.LabSectionService,
		r.LabMaterialService,
	)

	routes.NewCoreMaterialSubmissionRoutes(
		coreRouter,
		r.MaterialService,
		r.SubmissionService,
		r.LabSectionService,
	)
}
