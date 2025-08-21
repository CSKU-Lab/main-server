package rest

import (
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/gofiber/fiber/v2"
)

type AdminRouter struct {
	Router           fiber.Router
	UserService      services.UserService
	UserGroupService services.UserGroupService
	SemesterService  services.SemesterService
	CourseService    services.CourseService
}

func NewAdminRouter(r *AdminRouter) {
	adminRouter := r.Router.Group("/admin", middlewares.AdminMiddleware)

	routes.NewAdminUserRoutes(adminRouter, r.UserService)
	routes.NewAdminSemesterRoutes(adminRouter, r.SemesterService)
	routes.NewAdminCourseRoutes(adminRouter, r.CourseService)
	routes.NewAdminUserGroupRoutes(adminRouter, r.UserGroupService)

}
