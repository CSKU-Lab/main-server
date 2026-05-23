package providers

import (
	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest"
	"github.com/CSKU-Lab/queue"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"go.uber.org/zap"
)

func NewFiberApp(
	cfg *configs.Config,
	logger *zap.SugaredLogger,
	errMiddleware *middlewares.ErrorHandlerMiddleware,
	userService services.UserService,
	refreshTokenService services.RefreshTokenService,
	userGroupService services.UserGroupService,
	courseService services.CourseService,
	courseEnrollmentService services.CourseEnrollmentService,
	semesterService services.SemesterService,
	sectionLogService services.SectionLogService,
	sectionService services.SectionService,
	sectionStudentService services.SectionStudentService,
	tagService services.TagService,
	materialAssetService services.MaterialAssetService,
	labService services.LabService,
	labSectionService services.LabSectionService,
	labMaterialService services.LabMaterialService,
	defaultLabService services.DefaultLabService,
	affectedEntitiesService services.AffectedEntitiesService,
	submissionService services.SubmissionService,
	sidebarService services.SidebarService,
	gradebookExportService services.GradebookExportService,
	materialService services.MaterialService,
	searchService services.SearchService,
	permissionService permission.Service,
	configGRPCClient configPB.ConfigServiceClient,
	q queue.Queue,
	rClient pubsub.PubSub,
	playgroundHandler *rest.PlaygroundHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errMiddleware.ErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10 MB
	})

	app.Use(middlewares.OtelMiddleware())
	app.Use(middlewares.RequestLoggerMiddleware(logger))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FRONTEND_URL},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	api := app.Group("/api/v1")

	rest.NewPlaygroundRouter(api, playgroundHandler)

	rest.NewAuthRouter(api, cfg, userService, refreshTokenService)

	protectedApi := api.Group("/", middlewares.ProtectedRouteMiddleware(cfg.JWTSecret))

	rest.NewAdminRouter(&rest.AdminRouter{
		Router:            protectedApi,
		UserService:       userService,
		CourseService:     courseService,
		UserGroupService:  userGroupService,
		PermissionService: permissionService,
	})

	rest.NewCMSRouter(&rest.CMSRouter{
		Router:                  protectedApi,
		UserService:             userService,
		SemesterService:         semesterService,
		CourseService:           courseService,
		SectionService:          sectionService,
		MaterialService:         materialService,
		AffectedEntitiesService: affectedEntitiesService,
		LabService:              labService,
		LabSectionService:       labSectionService,
		LabMaterialService:      labMaterialService,
		DefaultLabService:       defaultLabService,
		SectionLogService:       sectionLogService,
		MaterialAssetService:    materialAssetService,
		ConfigGRPCClient:        configGRPCClient,
		SubmissionService:       submissionService,
		GradebookExportService:  gradebookExportService,
		SearchService:           searchService,
		Queue:                   q,
		PermissionService:       permissionService,
		TagService:              tagService,
	})

	rest.NewCoreRouter(&rest.CoreRouter{
		Router:                  protectedApi,
		SectionService:          sectionService,
		LabSectionService:       labSectionService,
		LabService:              labService,
		SectionStudentService:   sectionStudentService,
		LabMaterialService:      labMaterialService,
		CourseService:           courseService,
		CourseEnrollmentService: courseEnrollmentService,
		DefaultLabService:       defaultLabService,
		SidebarService:          sidebarService,
		MaterialService:         materialService,
		SubmissionService:       submissionService,
		SearchService:           searchService,
		PubSub:                  rClient,
		PermissionService:       permissionService,
		Secret:                  cfg.JWTSecret,
	})

	return app
}
