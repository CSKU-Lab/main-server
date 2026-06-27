package rest

import (
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/gofiber/fiber/v3"
)

type AdminRouter struct {
	Router                fiber.Router
	UserService           services.UserService
	UserGroupService      services.UserGroupService
	CourseService         services.CourseService
	PermissionService     permission.Service
	SystemSettingsService services.SystemSettingsService
	AnalyticsService      services.AnalyticsService
}

func NewAdminRouter(r *AdminRouter) {
	adminRouter := r.Router.Group("/admin")

	routes.NewAdminUserRoutes(adminRouter, r.UserService)
	routes.NewAdminUserGroupRoutes(adminRouter, r.UserGroupService)
	routes.NewAdminSettingsRoutes(adminRouter, r.SystemSettingsService)
	routes.NewAdminAnalyticsRoutes(adminRouter, r.AnalyticsService)
}
