package rest

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/gofiber/fiber/v3"
)

type CoreRouter struct {
	Router                  fiber.Router
	SectionService          services.SectionService
	LabSectionService       services.LabSectionService
	LabService              services.LabService
	SectionStudentService   services.SectionStudentService
	LabMaterialService      services.LabMaterialService
	CourseService           services.CourseService
	CourseEnrollmentService services.CourseEnrollmentService
	DefaultLabService       services.DefaultLabService
	SidebarService          services.SidebarService
	MaterialService         services.MaterialService
	SubmissionService       services.SubmissionService
	SearchService           services.SearchService
	PubSub                  pubsub.PubSub
	PermissionService          permission.Service
	TypingSubmissionRepository repositories.TypingSubmissionRepository
	InputSubmissionService     services.InputSubmissionService
	Secret                     string
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
		r.SubmissionService,
		r.PermissionService,
	)
	routes.NewCoreLabRoute(
		coreRouter,
		r.SectionService,
		r.LabSectionService,
		r.LabService,
		r.SectionStudentService,
		r.LabMaterialService,
		r.SubmissionService,
		r.PermissionService,
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
		r.PermissionService,
	)

	routes.NewCoreMaterialSubmissionRoutes(
		coreRouter,
		r.MaterialService,
		r.SubmissionService,
		r.LabSectionService,
		r.PermissionService,
	)

	routes.NewCoreCourseRoute(
		coreRouter,
		r.CourseService,
		r.CourseEnrollmentService,
		r.DefaultLabService,
		r.LabMaterialService,
		r.SectionService,
	)

	routes.NewCoreTypingRoute(
		coreRouter,
		r.MaterialService,
		r.CourseEnrollmentService,
		r.LabSectionService,
		r.TypingSubmissionRepository,
		r.Secret,
	)

	routes.NewCoreInputRoutes(
		coreRouter,
		r.InputSubmissionService,
	)

	routes.NewCoreSearchRoutes(
		coreRouter,
		r.SearchService,
	)

	routes.NewLspTokenRoute(coreRouter, r.Secret)
}
