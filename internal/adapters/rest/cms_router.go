package rest

import (
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/gofiber/fiber/v2"
)

type CMSRouter struct {
	Router          fiber.Router
	UserService     services.UserService
	SemesterService services.SemesterService
	CourseService   services.CourseService
	SectionService  services.SectionService
}

func NewCMSRouter(r *CMSRouter) {
	cmsRouter := r.Router.Group("/cms", middlewares.AdminMiddleware)

	routes.NewCmsSectionRoutes(cmsRouter, r.SectionService, r.SemesterService)
}
