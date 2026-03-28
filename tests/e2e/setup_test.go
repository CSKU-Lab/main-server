//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/registrables"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest"
	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

// TestSuite is the base E2E test suite that provides:
// - Database connection with transaction rollback for isolation
// - HTTP test server
// - Test data fixtures
// - Authentication helpers
type TestSuite struct {
	suite.Suite
	DB       *sqlx.DB
	App      *fiber.App
	Ctx      context.Context
	TestUser *TestUserFixture
	TestData *TestDataManager
}

// TestUserFixture holds authentication tokens for different test users
type TestUserFixture struct {
	Admin      *UserTokens
	Instructor *UserTokens
	Student    *UserTokens
	Student2   *UserTokens
}

// UserTokens holds authentication tokens for a user
type UserTokens struct {
	UserID       string
	Username     string
	AccessToken  string
	RefreshToken string
	Roles        []string
}

// TestDataManager manages test data lifecycle
type TestDataManager struct {
	CourseID   string
	SectionID  string
	LabID      string
	MaterialID string
}

// getDatabaseURL constructs the database connection string from environment variables
func getDatabaseURL() string {
	host := getEnv("PGHOST", "localhost")
	port := getEnv("PGPORT", "5432")
	user := getEnv("PGUSER", "cs_pg_user")
	password := getEnv("PGPASSWORD", "cs_pg_password")
	dbname := getEnv("PGDATABASE", "main-server")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupSuite runs once before all tests in the suite
func (s *TestSuite) SetupSuite() {
	s.Ctx = context.Background()

	// Connect to PostgreSQL
	databaseURL := getDatabaseURL()
	db, err := sqlx.Connect("postgres", databaseURL)
	s.Require().NoError(err, "Failed to connect to PostgreSQL database")

	// Verify connection
	err = db.Ping()
	s.Require().NoError(err, "Failed to ping PostgreSQL database")

	s.DB = db

	// Initialize the Fiber app with all routes and middleware
	s.initTestApp()
	s.Require().NotNil(s.App, "Failed to initialize Fiber app")

	// Initialize test data manager
	s.TestData = &TestDataManager{}

	// Create test users and get tokens
	s.setupTestUsers()
}

// TearDownSuite runs once after all tests in the suite
func (s *TestSuite) TearDownSuite() {
	// Clean up all test data
	s.cleanupAllTestData()

	if s.DB != nil {
		s.DB.Close()
	}
}

// SetupTest runs before each test
func (s *TestSuite) SetupTest() {
	// Individual tests should handle their own setup
	// This ensures test isolation
}

// TearDownTest runs after each test
func (s *TestSuite) TearDownTest() {
	// Clean up test-specific data
	s.cleanupTestData()
}

// setupTestUsers creates test users with different roles and obtains JWT tokens
func (s *TestSuite) setupTestUsers() {
	s.TestUser = &TestUserFixture{}

	// Create admin user and get actual username
	adminID := s.CreateTestUser("admin", []string{"admin"})
	var adminUsername string
	s.DB.Get(&adminUsername, "SELECT username FROM users WHERE id = $1", adminID)
	s.TestUser.Admin = &UserTokens{
		UserID:   adminID,
		Username: adminUsername,
		Roles:    []string{"admin"},
	}

	// Create instructor user and get actual username
	instructorID := s.CreateTestUser("instructor", []string{"instructor"})
	var instructorUsername string
	s.DB.Get(&instructorUsername, "SELECT username FROM users WHERE id = $1", instructorID)
	s.TestUser.Instructor = &UserTokens{
		UserID:   instructorID,
		Username: instructorUsername,
		Roles:    []string{"instructor"},
	}

	// Create student user and get actual username
	studentID := s.CreateTestUser("student", []string{"student"})
	var studentUsername string
	s.DB.Get(&studentUsername, "SELECT username FROM users WHERE id = $1", studentID)
	s.TestUser.Student = &UserTokens{
		UserID:   studentID,
		Username: studentUsername,
		Roles:    []string{"student"},
	}

	// Create second student user and get actual username
	student2ID := s.CreateTestUser("student2", []string{"student"})
	var student2Username string
	s.DB.Get(&student2Username, "SELECT username FROM users WHERE id = $1", student2ID)
	s.TestUser.Student2 = &UserTokens{
		UserID:   student2ID,
		Username: student2Username,
		Roles:    []string{"student"},
	}
}

// cleanupAllTestData removes all test data created during the test suite
func (s *TestSuite) cleanupAllTestData() {
	// Delete in reverse order of dependencies
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id IN (SELECT id FROM submissions WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'test_%'))")
	s.DB.Exec("DELETE FROM submissions WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'test_%')")
	s.DB.Exec("DELETE FROM section_students WHERE student_id IN (SELECT id FROM users WHERE username LIKE 'test_%')")
	s.DB.Exec("DELETE FROM section_instructors WHERE instructor_id IN (SELECT id FROM users WHERE username LIKE 'test_%')")
	s.DB.Exec("DELETE FROM course_creators WHERE creator_id IN (SELECT id FROM users WHERE username LIKE 'test_%')")
	s.DB.Exec("DELETE FROM lab_materials WHERE lab_id IN (SELECT id FROM labs WHERE display_name LIKE 'E2E Test%')")
	s.DB.Exec("DELETE FROM code_materials WHERE material_id IN (SELECT id FROM materials WHERE name LIKE 'E2E Test%')")
	s.DB.Exec("DELETE FROM materials WHERE name LIKE 'E2E Test%'")
	s.DB.Exec("DELETE FROM lab_sections WHERE lab_id IN (SELECT id FROM labs WHERE display_name LIKE 'E2E Test%')")
	s.DB.Exec("DELETE FROM default_labs WHERE lab_id IN (SELECT id FROM labs WHERE display_name LIKE 'E2E Test%')")
	s.DB.Exec("DELETE FROM labs WHERE display_name LIKE 'E2E Test%'")
	s.DB.Exec("DELETE FROM section_logs WHERE section_id IN (SELECT id FROM sections WHERE name LIKE 'E2E Test%')")
	s.DB.Exec("DELETE FROM sections WHERE name LIKE 'E2E Test%'")
	s.DB.Exec("DELETE FROM semesters WHERE name LIKE 'E2E Test%'")
	s.DB.Exec("DELETE FROM courses WHERE name LIKE 'E2E Test%'")
	s.DB.Exec("DELETE FROM user_passwords WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'test_%')")
	s.DB.Exec("DELETE FROM user_refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'test_%')")
	s.DB.Exec("DELETE FROM users WHERE username LIKE 'test_%'")
}

// cleanupTestData removes test data for a specific test
func (s *TestSuite) cleanupTestData() {
	// This can be overridden by individual tests
}

// GetJWTSecret returns the JWT secret from environment
func (s *TestSuite) GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "test-jwt-secret-for-e2e-tests"
	}
	return secret
}

// GetRefreshSecret returns the JWT refresh secret from environment
func (s *TestSuite) GetRefreshSecret() string {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		return "test-refresh-secret-for-e2e-tests"
	}
	return secret
}

// WaitForDatabaseConnection waits for the database to be available
func WaitForDatabaseConnection(t *testing.T) *sqlx.DB {
	databaseURL := getDatabaseURL()
	var db *sqlx.DB
	var err error

	// Retry connection up to 5 times with 1 second delay
	for i := 0; i < 5; i++ {
		db, err = sqlx.Connect("postgres", databaseURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				return db
			}
		}
		time.Sleep(1 * time.Second)
	}

	t.Fatalf("Failed to connect to database after 5 attempts: %v", err)
	return nil
}

// CheckE2EEnabled checks if E2E tests should run
func CheckE2EEnabled(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "true" && os.Getenv("CI") != "true" {
		t.Skip("Skipping E2E tests. Set RUN_E2E_TESTS=true to run.")
	}
}

// initTestApp initializes the Fiber app with all routes and middleware for testing
func (s *TestSuite) initTestApp() {
	// Create test configuration
	config := &configs.Config{
		ApiURL:           getEnv("API_URL", "http://localhost:8080"),
		DatabaseURL:      getDatabaseURL(),
		Port:             getEnv("PORT", "8080"),
		JWTSecret:        s.GetJWTSecret(),
		JWTRefreshSecret: s.GetRefreshSecret(),
		DevMode:          true, // Enable dev mode for testing
		FRONTEND_URL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		COOKIE_DOMAIN:    getEnv("COOKIE_DOMAIN", "localhost"),
		// gRPC and other service URLs - use defaults or env vars
		GRADER_SERVER_URL: getEnv("GRADER_SERVER_URL", "localhost:8083"),
		CONFIG_SERVER_URL: getEnv("CONFIG_SERVER_URL", "localhost:8081"),
		TASK_SERVER_URL:   getEnv("TASK_SERVER_URL", "localhost:8082"),
		RBMQ_SERVER_URL:   getEnv("RBMQ_SERVER_URL", "amqp://guest:guest@localhost:5672"),
		REDIS_SERVER_URL:  getEnv("REDIS_SERVER_URL", "localhost:6379"),
		// S3/MinIO config
		S3_AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", "minioadmin"),
		S3_SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", "minioadmin"),
		S3_UseSSL:          getEnv("S3_USE_SSL", "false") == "true",
		S3_Endpoint:        getEnv("S3_ENDPOINT", "localhost:9000"),
		S3_Frontend_URL:    getEnv("S3_FRONTEND_URL", "http://localhost:9000"),
		S3_Bucket:          getEnv("S3_BUCKET", "main-server"),
	}

	// Create error handler middleware
	errHandlerMiddleware := middlewares.NewErrorHandlerMiddleware(config)

	// Create repositories
	uowRepo := sqlxAdapter.NewUoWRepository(context.Background(), s.DB)

	userRepo := sqlxAdapter.NewUserRepository(s.DB)
	userPasswordRepo := sqlxAdapter.NewUserPasswordRepository(s.DB)
	userGroupRepo := sqlxAdapter.NewUserGroupRepository(s.DB)

	userService := services.NewUserService(userRepo, userPasswordRepo, userGroupRepo, uowRepo)
	userGroupService := services.NewUserGroupService(userGroupRepo)

	refreshTokenRepo := sqlxAdapter.NewSQLxRefreshTokenRepository(s.DB)
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo)

	courseRepo := sqlxAdapter.NewCourseRepository(s.DB)
	courseCreatorRepo := sqlxAdapter.NewCourseCreatorRepository(s.DB)
	courseService := services.NewCourseService(courseRepo, courseCreatorRepo, uowRepo)

	sectionRepo := sqlxAdapter.NewSectionRepository(s.DB)
	sectionInstructorRepo := sqlxAdapter.NewSectionInstructorRepository(s.DB)
	sectionStudentRepo := sqlxAdapter.NewSectionStudentRepository(s.DB)
	sectionLogRepo := sqlxAdapter.NewSectionLogRepository(s.DB)

	sectionStudentService := services.NewSectionStudentService(sectionStudentRepo, sectionRepo, userRepo)

	semesterRepo := sqlxAdapter.NewSqlxSemesterRepository(s.DB)
	semesterService := services.NewSemesterService(semesterRepo, sectionRepo, courseRepo)

	sectionLogService := services.NewSectionLogService(sectionLogRepo)

	// Create a minimal section service without external dependencies for testing
	sectionService := services.NewSectionService(config, sectionRepo, uowRepo, courseRepo, sectionInstructorRepo, sectionStudentRepo, nil, userRepo, semesterRepo, sectionLogService)

	materialRepo := sqlxAdapter.NewMaterialRepository(s.DB)
	readMaterialTagRepo := sqlxAdapter.NewReadMaterialTagRepository(s.DB)
	codeMaterialRepo := sqlxAdapter.NewCodeMaterialRepository(s.DB)

	materialRegistry := registries.NewMaterialRegistry()
	// Register code material handler - needed for submission tests
	taskStub := NewTaskServiceStub()
	configStub := NewConfigServiceStub()
	codeMaterial := registrables.NewCodeMaterial(codeMaterialRepo, taskStub, configStub)
	materialRegistry.Register("code", codeMaterial)

	labRepo := sqlxAdapter.NewSqlxLabRepository(s.DB)
	labService := services.NewLabService(labRepo, courseRepo, uowRepo)

	labSectionRepo := sqlxAdapter.NewSqlxLabSectionRepository(s.DB)
	labSectionService := services.NewLabSectionService(labSectionRepo, uowRepo, labRepo, sectionRepo, sectionStudentRepo)

	labMaterialRepo := sqlxAdapter.NewSqlxLabMaterialRepository(s.DB)
	labMaterialService := services.NewLabMaterialService(labMaterialRepo, uowRepo, labRepo, materialRepo, readMaterialTagRepo)

	defaultLabRepo := sqlxAdapter.NewSqlxDefaultLabRepository(s.DB)
	defaultLabService := services.NewDefaultLabService(defaultLabRepo, uowRepo, courseRepo, labRepo)

	affectedEntitiesFactory := registries.NewAffectedEntityFactory()
	deletedCourseAffected := registrables.NewDeletedCourseAffected(courseRepo, sectionRepo)
	deletedSemesterAffected := registrables.NewDeletedSemesterAffected(semesterRepo, sectionRepo, courseRepo)
	deletedSectionAffected := registrables.NewDeletedSectionAffected(sectionStudentRepo)
	deletedLabAffected := registrables.NewDeletedLabAffected(labRepo, labSectionRepo, labMaterialRepo, defaultLabRepo, uowRepo)
	deletedLabSectionAffected := registrables.NewDeletedLabSectionAffected()

	affectedEntitiesFactory.Register("course", deletedCourseAffected)
	affectedEntitiesFactory.Register("semester", deletedSemesterAffected)
	affectedEntitiesFactory.Register("section", deletedSectionAffected)
	affectedEntitiesFactory.Register("lab", deletedLabAffected)
	affectedEntitiesFactory.Register("lab_section", deletedLabSectionAffected)

	affectedEntitiesService := services.NewAffectedEntitiesService(affectedEntitiesFactory)

	sidebarService := services.NewSidebarService(courseRepo, sectionStudentRepo, labSectionRepo, labMaterialRepo)

	submissionRepo := sqlxAdapter.NewSubmissionRepository(s.DB)
	codeSubmissionRepo := sqlxAdapter.NewCodeSubmission(s.DB)

	submissionRegistry := registries.NewSubmission()
	// Register code submission handler - needed for submission tests
	codeSubmissionRegistrable := registrables.NewCodeSubmission(codeSubmissionRepo, codeMaterialRepo, submissionRepo, taskStub)
	submissionRegistry.Register("code", codeSubmissionRegistrable)

	// Create submission service without external dependencies (PubSub, Queue, gRPC)
	submissionService := services.NewSubmissionService(&services.SubmissionServiceArgs{
		SubmissionRepository:     submissionRepo,
		MaterialRepository:       materialRepo,
		UowRepository:            uowRepo,
		SubmissionRegistry:       submissionRegistry,
		SectionStudentRepository: sectionStudentRepo,
		UserRepository:           userRepo,
		MaterialRegistry:         materialRegistry,
		SectionRepository:        sectionRepo,
		LabSectionRepository:     labSectionRepo,
		LabMaterialRepository:    labMaterialRepo,
		PubSub:                   nil, // No PubSub in tests
	})

	gradebookExportService := services.NewGradebookExportService(submissionService)

	materialService := services.NewMaterialService(materialRepo, submissionRepo, readMaterialTagRepo, uowRepo, userRepo, materialRegistry)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: errHandlerMiddleware.ErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10 MB
	})

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Create API group
	api := app.Group("/api/v1")

	// Setup auth routes
	rest.NewAuthRouter(api, config, userService, refreshTokenService)

	// Setup protected routes
	protectedApi := api.Group("/", middlewares.ProtectedRouteMiddleware(config.JWTSecret))

	// Setup admin routes
	rest.NewAdminRouter(&rest.AdminRouter{
		Router:           protectedApi,
		UserService:      userService,
		UserGroupService: userGroupService,
		CourseService:    courseService,
	})

	// Setup CMS routes
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
		MaterialAssetService:    nil, // No MinIO in tests
		ConfigGRPCClient:        nil, // No gRPC in tests
		SubmissionService:       submissionService,
		GradebookExportService:  gradebookExportService,
		Queue:                   nil, // No RabbitMQ in tests
	})

	// Setup core routes
	rest.NewCoreRouter(&rest.CoreRouter{
		Router:                protectedApi,
		SectionService:        sectionService,
		LabSectionService:     labSectionService,
		LabService:            labService,
		SectionStudentService: sectionStudentService,
		LabMaterialService:    labMaterialService,
		CourseService:         courseService,
		SidebarService:        sidebarService,
		MaterialService:       materialService,
		SubmissionService:     submissionService,
		SubmissionRepo:        submissionRepo,
		PubSub:                nil, // No PubSub in tests
	})

	s.App = app
}
