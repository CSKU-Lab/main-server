package providers

import (
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/domain/services"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/adapters/ratelimit"
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
	typingExportService services.TypingExportService,
	materialService services.MaterialService,
	searchService services.SearchService,
	permissionService permission.Service,
	systemSettingsService services.SystemSettingsService,
	analyticsService services.AnalyticsService,
	configGRPCClient configPB.ConfigServiceClient,
	q queue.Queue,
	rClient pubsub.PubSub,
	rateLimiter ratelimit.RateLimiter,
	playgroundHandler *rest.PlaygroundHandler,
	typingSubRepo repositories.TypingSubmissionRepository,
	inputSubmissionService services.InputSubmissionService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errMiddleware.ErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10 MB
		// Real client IP sits in X-Forwarded-For set by Traefik. Trust only the
		// in-cluster proxy hop (private + loopback ranges) so c.IP() reads XFF
		// from Traefik but never an attacker-spoofed header from an untrusted source.
		ProxyHeader: fiber.HeaderXForwardedFor,
		TrustProxy:  true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Private:  true,
			Loopback: true,
		},
	})

	app.Use(middlewares.OtelMiddleware())
	app.Use(middlewares.RequestLoggerMiddleware(logger))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FRONTEND_URL},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	api := app.Group("/api/v1")

	rest.NewAuthRouter(api, cfg, userService, refreshTokenService)

	rest.NewInternalRouter(&rest.InternalRouter{
		Router:          api,
		InternalToken:   cfg.InternalToken,
		CourseService:   courseService,
		SectionService:  sectionService,
		MaterialService: materialService,
		Queue:           q,
	})

	protectedApi := api.Group("/", middlewares.ProtectedRouteMiddleware(cfg.JWTSecret))

	rest.NewPlaygroundRouter(protectedApi, playgroundHandler)

	rest.NewAdminRouter(&rest.AdminRouter{
		Router:                protectedApi,
		UserService:           userService,
		CourseService:         courseService,
		UserGroupService:      userGroupService,
		PermissionService:     permissionService,
		SystemSettingsService: systemSettingsService,
		AnalyticsService:      analyticsService,
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
		TypingExportService:     typingExportService,
		SearchService:           searchService,
		Queue:                   q,
		PermissionService:       permissionService,
		TagService:              tagService,
		AnalyticsService:        analyticsService,
		InputSubmissionService:  inputSubmissionService,
	})

	coreApi := protectedApi.Group("/", middlewares.RateLimitMiddleware(rateLimiter, 300, time.Minute))

	rest.NewCoreRouter(&rest.CoreRouter{
		Router:                     coreApi,
		SectionService:             sectionService,
		LabSectionService:          labSectionService,
		LabService:                 labService,
		SectionStudentService:      sectionStudentService,
		LabMaterialService:         labMaterialService,
		CourseService:              courseService,
		CourseEnrollmentService:    courseEnrollmentService,
		DefaultLabService:          defaultLabService,
		SidebarService:             sidebarService,
		MaterialService:            materialService,
		SubmissionService:          submissionService,
		SearchService:              searchService,
		PubSub:                     rClient,
		PermissionService:          permissionService,
		TypingSubmissionRepository: typingSubRepo,
		InputSubmissionService:     inputSubmissionService,
		Secret:                     cfg.JWTSecret,
	})

	return app
}
