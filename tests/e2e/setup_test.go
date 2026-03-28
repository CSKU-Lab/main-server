//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
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

	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	s.TestUser.Admin = &UserTokens{
		UserID:   adminID,
		Username: "test_admin",
		Roles:    []string{"admin"},
	}

	// Create instructor user
	instructorID := s.CreateTestUser("instructor", []string{"instructor"})
	s.TestUser.Instructor = &UserTokens{
		UserID:   instructorID,
		Username: "test_instructor",
		Roles:    []string{"instructor"},
	}

	// Create student user
	studentID := s.CreateTestUser("student", []string{"student"})
	s.TestUser.Student = &UserTokens{
		UserID:   studentID,
		Username: "test_student",
		Roles:    []string{"student"},
	}

	// Create second student user
	student2ID := s.CreateTestUser("student2", []string{"student"})
	s.TestUser.Student2 = &UserTokens{
		UserID:   student2ID,
		Username: "test_student2",
		Roles:    []string{"student"},
	}
}

// cleanupAllTestData removes all test data created during the test suite
func (s *TestSuite) cleanupAllTestData() {
	// Delete in reverse order of dependencies
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id IN (SELECT id FROM submissions WHERE user_id LIKE 'e2e_test_%')")
	s.DB.Exec("DELETE FROM submissions WHERE user_id LIKE 'e2e_test_%'")
	s.DB.Exec("DELETE FROM section_students WHERE student_id LIKE 'e2e_test_%'")
	s.DB.Exec("DELETE FROM section_instructors WHERE instructor_id LIKE 'e2e_test_%'")
	s.DB.Exec("DELETE FROM sections WHERE name LIKE 'E2E Test%'")
	s.DB.Exec("DELETE FROM course_creators WHERE course_id IN (SELECT id FROM courses WHERE name LIKE 'E2E Test%')")
	s.DB.Exec("DELETE FROM courses WHERE name LIKE 'E2E Test%'")
	s.DB.Exec("DELETE FROM users WHERE id LIKE 'e2e_test_%'")
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
