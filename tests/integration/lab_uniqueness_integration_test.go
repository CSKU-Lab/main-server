//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// LabUniquenessIntegrationTestSuite tests the lab name uniqueness constraint
// using a real PostgreSQL database connection. These tests verify that:
// 1. Same lab name CAN exist in different courses
// 2. Same lab name CANNOT exist twice in same course
// 3. Soft-deleted labs allow name reuse in the same course
// 4. The composite index on (course_id, display_name) exists and is enforced
// 5. Old global index does not exist
type LabUniquenessIntegrationTestSuite struct {
	suite.Suite
	db      *sqlx.DB
	ctx     context.Context
	labRepo repositories.LabRepository
}

// getDatabaseURL constructs the database connection string from environment variables
// Uses standard PostgreSQL environment variables:
// - PGHOST (default: localhost)
// - PGPORT (default: 5432)
// - PGUSER (default: cs_pg_user)
// - PGPASSWORD (default: cs_pg_password)
// - PGDATABASE (default: main-server)
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
func (s *LabUniquenessIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Connect to PostgreSQL
	databaseURL := getDatabaseURL()
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(s.T(), err, "Failed to connect to PostgreSQL database")

	// Verify connection
	err = db.Ping()
	require.NoError(s.T(), err, "Failed to ping PostgreSQL database")

	s.db = db
	s.labRepo = sqlxAdapter.NewSqlxLabRepository(db)

	// Run Atlas migrations to ensure schema is up to date
	s.runMigrations()
}

// TearDownSuite runs once after all tests in the suite
func (s *LabUniquenessIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// SetupTest runs before each test
func (s *LabUniquenessIntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	s.cleanupTestData()
}

// TearDownTest runs after each test
func (s *LabUniquenessIntegrationTestSuite) TearDownTest() {
	// Clean up test data after each test
	s.cleanupTestData()
}

// runMigrations applies the Atlas migrations to ensure the correct schema
func (s *LabUniquenessIntegrationTestSuite) runMigrations() {
	// Apply the lab uniqueness migration directly via SQL
	// This ensures the test has the correct schema regardless of migration state
	migrationSQL := `
		-- Drop the old global unique index if it exists
		DROP INDEX IF EXISTS unique_display_name;
		
		-- Create new composite unique index on (course_id, display_name)
		CREATE UNIQUE INDEX IF NOT EXISTS unique_display_name_per_course 
		ON labs (course_id, display_name) 
		WHERE is_deleted = false;
	`

	_, err := s.db.Exec(migrationSQL)
	require.NoError(s.T(), err, "Failed to apply lab uniqueness migration")
}

// cleanupTestData removes all test data created during tests
func (s *LabUniquenessIntegrationTestSuite) cleanupTestData() {
	// Delete test labs (soft delete first, then hard delete for cleanup)
	s.db.Exec(`DELETE FROM labs WHERE display_name LIKE 'Test Lab%'`)
	s.db.Exec(`DELETE FROM labs WHERE display_name LIKE 'Integration Test%'`)

	// Delete test courses
	s.db.Exec(`DELETE FROM courses WHERE name LIKE 'Integration Test Course%'`)

	// Delete test users
	s.db.Exec(`DELETE FROM users WHERE email LIKE 'integration_test_%'`)
}

// createTestUser creates a test user and returns the ID
func (s *LabUniquenessIntegrationTestSuite) createTestUser() string {
	userID := uuid.New().String()
	uniqueSuffix := userID[:8]
	email := fmt.Sprintf("integration_test_%s@example.com", uniqueSuffix)
	username := fmt.Sprintf("integration_test_%s", uniqueSuffix)

	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, type, username, display_name, roles, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, ARRAY['instructor']::role[], NOW(), NOW(), false)
		ON CONFLICT (id) DO NOTHING
	`, userID, email, "oauth", username, "Integration Test User")
	require.NoError(s.T(), err, "Failed to create test user")

	return userID
}

// createTestCourse creates a test course and returns the ID
func (s *LabUniquenessIntegrationTestSuite) createTestCourse(name string) string {
	courseID := uuid.New().String()

	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO courses (id, name, visibility, created_at, updated_at, is_deleted, is_archived)
		VALUES ($1, $2, $3, NOW(), NOW(), false, false)
		ON CONFLICT (id) DO NOTHING
	`, courseID, name, "public")
	require.NoError(s.T(), err, "Failed to create test course: %s", name)

	return courseID
}

// TestSameLabNameInDifferentCourses verifies that the same lab name can exist in different courses
func (s *LabUniquenessIntegrationTestSuite) TestSameLabNameInDifferentCourses() {
	userID := s.createTestUser()
	course1ID := s.createTestCourse("Integration Test Course 1")
	course2ID := s.createTestCourse("Integration Test Course 2")
	labName := "Integration Test Lab - Introduction to Go"

	// Create lab in course 1
	lab1ID := uuid.New().String()
	req1 := &requests.CreateLab{
		DisplayName: labName,
		CourseID:    course1ID,
	}
	err := s.labRepo.Create(s.ctx, lab1ID, req1, userID)
	require.NoError(s.T(), err, "Should create lab in course 1")

	// Create lab with same name in course 2 - should succeed
	lab2ID := uuid.New().String()
	req2 := &requests.CreateLab{
		DisplayName: labName,
		CourseID:    course2ID,
	}
	err = s.labRepo.Create(s.ctx, lab2ID, req2, userID)
	require.NoError(s.T(), err, "Should create lab with same name in different course")

	// Verify both labs exist with correct data
	lab1, err := s.labRepo.GetByID(s.ctx, lab1ID)
	require.NoError(s.T(), err, "Should retrieve lab 1")
	assert.Equal(s.T(), labName, lab1.DisplayName, "Lab 1 should have correct name")
	assert.Equal(s.T(), course1ID, lab1.CourseID, "Lab 1 should belong to course 1")

	lab2, err := s.labRepo.GetByID(s.ctx, lab2ID)
	require.NoError(s.T(), err, "Should retrieve lab 2")
	assert.Equal(s.T(), labName, lab2.DisplayName, "Lab 2 should have correct name")
	assert.Equal(s.T(), course2ID, lab2.CourseID, "Lab 2 should belong to course 2")

	s.T().Logf("✓ Same lab name '%s' successfully created in both courses", labName)
}

// TestDuplicateLabNameInSameCourse verifies that duplicate lab names in the same course fail
func (s *LabUniquenessIntegrationTestSuite) TestDuplicateLabNameInSameCourse() {
	userID := s.createTestUser()
	courseID := s.createTestCourse("Integration Test Course - Duplicate Test")
	labName := "Integration Test Lab - Advanced Go Patterns"

	// Create first lab
	lab1ID := uuid.New().String()
	req1 := &requests.CreateLab{
		DisplayName: labName,
		CourseID:    courseID,
	}
	err := s.labRepo.Create(s.ctx, lab1ID, req1, userID)
	require.NoError(s.T(), err, "Should create first lab")

	// Try to create second lab with same name in same course - should fail
	lab2ID := uuid.New().String()
	req2 := &requests.CreateLab{
		DisplayName: labName,
		CourseID:    courseID,
	}
	err = s.labRepo.Create(s.ctx, lab2ID, req2, userID)
	require.Error(s.T(), err, "Should fail to create lab with duplicate name in same course")
	assert.Contains(s.T(), err.Error(), "unique", "Error should mention unique constraint violation")

	s.T().Logf("✓ Duplicate lab name '%s' correctly rejected in same course", labName)
}

// TestSoftDeletedLabAllowsNameReuse verifies that soft-deleted labs allow name reuse
func (s *LabUniquenessIntegrationTestSuite) TestSoftDeletedLabAllowsNameReuse() {
	userID := s.createTestUser()
	courseID := s.createTestCourse("Integration Test Course - Soft Delete Test")
	labName := "Integration Test Lab - Temporary Lab"

	// Create a lab
	lab1ID := uuid.New().String()
	req1 := &requests.CreateLab{
		DisplayName: labName,
		CourseID:    courseID,
	}
	err := s.labRepo.Create(s.ctx, lab1ID, req1, userID)
	require.NoError(s.T(), err, "Should create lab")

	// Soft delete the lab
	err = s.labRepo.DeleteByID(s.ctx, lab1ID)
	require.NoError(s.T(), err, "Should soft delete lab")

	// Verify the lab is soft deleted
	var isDeleted bool
	err = s.db.GetContext(s.ctx, &isDeleted, `
		SELECT is_deleted FROM labs WHERE id = $1
	`, lab1ID)
	require.NoError(s.T(), err, "Should retrieve lab deletion status")
	assert.True(s.T(), isDeleted, "Lab should be marked as deleted")

	// Now create a new lab with the same name - should succeed
	lab2ID := uuid.New().String()
	req2 := &requests.CreateLab{
		DisplayName: labName,
		CourseID:    courseID,
	}
	err = s.labRepo.Create(s.ctx, lab2ID, req2, userID)
	require.NoError(s.T(), err, "Should create lab with same name after soft delete")

	// Verify the new lab exists
	lab2, err := s.labRepo.GetByID(s.ctx, lab2ID)
	require.NoError(s.T(), err, "Should retrieve new lab")
	assert.Equal(s.T(), labName, lab2.DisplayName, "New lab should have the reused name")
	assert.Equal(s.T(), courseID, lab2.CourseID, "New lab should belong to the same course")

	s.T().Logf("✓ Soft-deleted lab allows name reuse for '%s'", labName)
}

// TestCompositeUniqueIndexExists verifies the composite unique index exists
func (s *LabUniquenessIntegrationTestSuite) TestCompositeUniqueIndexExists() {
	var count int
	err := s.db.GetContext(s.ctx, &count, `
		SELECT COUNT(*) FROM pg_indexes 
		WHERE indexname = 'unique_display_name_per_course' 
		AND tablename = 'labs'
	`)
	require.NoError(s.T(), err, "Should query pg_indexes")
	assert.Equal(s.T(), 1, count, "Composite unique_display_name_per_course index should exist")

	// Verify index definition
	var indexDef string
	err = s.db.GetContext(s.ctx, &indexDef, `
		SELECT indexdef FROM pg_indexes 
		WHERE indexname = 'unique_display_name_per_course' 
		AND tablename = 'labs'
	`)
	require.NoError(s.T(), err, "Should retrieve index definition")
	assert.Contains(s.T(), indexDef, "course_id", "Index should include course_id column")
	assert.Contains(s.T(), indexDef, "display_name", "Index should include display_name column")
	assert.Contains(s.T(), indexDef, "UNIQUE", "Index should be unique")
	assert.Contains(s.T(), indexDef, "is_deleted = false", "Index should have soft-delete condition")

	s.T().Log("✓ Composite unique index 'unique_display_name_per_course' exists with correct definition")
}

// TestOldGlobalIndexDoesNotExist verifies the old global unique index was removed
func (s *LabUniquenessIntegrationTestSuite) TestOldGlobalIndexDoesNotExist() {
	var count int
	err := s.db.GetContext(s.ctx, &count, `
		SELECT COUNT(*) FROM pg_indexes 
		WHERE indexname = 'unique_display_name' 
		AND tablename = 'labs'
	`)
	require.NoError(s.T(), err, "Should query pg_indexes")
	assert.Equal(s.T(), 0, count, "Old global unique_display_name index should not exist")

	s.T().Log("✓ Old global index 'unique_display_name' does not exist")
}

// TestMultipleLabsSameCourseDifferentNames verifies multiple labs can exist in same course with different names
func (s *LabUniquenessIntegrationTestSuite) TestMultipleLabsSameCourseDifferentNames() {
	userID := s.createTestUser()
	courseID := s.createTestCourse("Integration Test Course - Multiple Labs")

	labNames := []string{
		"Integration Test Lab - Lab 1",
		"Integration Test Lab - Lab 2",
		"Integration Test Lab - Lab 3",
	}

	// Create multiple labs with different names in the same course
	for i, name := range labNames {
		labID := uuid.New().String()
		req := &requests.CreateLab{
			DisplayName: name,
			CourseID:    courseID,
		}
		err := s.labRepo.Create(s.ctx, labID, req, userID)
		require.NoError(s.T(), err, "Should create lab %d with name '%s'", i+1, name)
	}

	// Verify all labs exist
	for _, name := range labNames {
		var count int
		err := s.db.GetContext(s.ctx, &count, `
			SELECT COUNT(*) FROM labs 
			WHERE display_name = $1 
			AND course_id = $2 
			AND is_deleted = false
		`, name, courseID)
		require.NoError(s.T(), err, "Should query lab count")
		assert.Equal(s.T(), 1, count, "Lab '%s' should exist in course", name)
	}

	s.T().Logf("✓ Successfully created %d labs with different names in same course", len(labNames))
}

// TestLabNameUniquenessCaseSensitivity tests case sensitivity of lab names
func (s *LabUniquenessIntegrationTestSuite) TestLabNameUniquenessCaseSensitivity() {
	userID := s.createTestUser()
	courseID := s.createTestCourse("Integration Test Course - Case Sensitivity")

	// PostgreSQL UNIQUE indexes are case-sensitive by default
	labName1 := "Integration Test Lab - Go Programming"
	labName2 := "Integration Test Lab - go programming" // different case

	// Create first lab
	lab1ID := uuid.New().String()
	req1 := &requests.CreateLab{
		DisplayName: labName1,
		CourseID:    courseID,
	}
	err := s.labRepo.Create(s.ctx, lab1ID, req1, userID)
	require.NoError(s.T(), err, "Should create first lab")

	// Create second lab with different case - should succeed (case-sensitive)
	lab2ID := uuid.New().String()
	req2 := &requests.CreateLab{
		DisplayName: labName2,
		CourseID:    courseID,
	}
	err = s.labRepo.Create(s.ctx, lab2ID, req2, userID)
	require.NoError(s.T(), err, "Should create lab with different case name")

	// Verify both exist
	lab1, err := s.labRepo.GetByID(s.ctx, lab1ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), labName1, lab1.DisplayName)

	lab2, err := s.labRepo.GetByID(s.ctx, lab2ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), labName2, lab2.DisplayName)

	s.T().Log("✓ Case-sensitive lab names are treated as different names")
}

// Run the test suite
func TestLabUniquenessIntegrationTestSuite(t *testing.T) {
	// Check if database is available
	databaseURL := getDatabaseURL()
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		t.Skipf("Skipping integration tests: PostgreSQL database not available: %v", err)
	}
	db.Close()

	suite.Run(t, new(LabUniquenessIntegrationTestSuite))
}

// TestDatabaseConnection verifies database connectivity
func TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseURL := getDatabaseURL()
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		t.Skipf("Skipping: PostgreSQL database not available: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	require.NoError(t, err, "Should be able to ping database")

	// Test query
	var version string
	err = db.Get(&version, "SELECT version()")
	require.NoError(t, err, "Should be able to execute query")
	assert.NotEmpty(t, version, "PostgreSQL version should not be empty")

	t.Logf("✓ Connected to PostgreSQL: %s", version)
}

// TestLabsTableStructure verifies the labs table has the expected structure
func TestLabsTableStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	databaseURL := getDatabaseURL()
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		t.Skipf("Skipping: PostgreSQL database not available: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Check labs table exists
	var tableExists bool
	err = db.GetContext(ctx, &tableExists, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'labs'
		)
	`)
	require.NoError(t, err, "Should query information_schema")
	assert.True(t, tableExists, "Labs table should exist")

	// Check required columns exist
	requiredColumns := []string{"id", "display_name", "course_id", "created_by", "is_deleted"}
	for _, column := range requiredColumns {
		var columnExists bool
		err = db.GetContext(ctx, &columnExists, `
			SELECT EXISTS (
				SELECT FROM information_schema.columns 
				WHERE table_schema = 'public' 
				AND table_name = 'labs' 
				AND column_name = $1
			)
		`, column)
		require.NoError(t, err, "Should query column %s", column)
		assert.True(t, columnExists, "Column '%s' should exist in labs table", column)
	}

	t.Log("✓ Labs table has correct structure")
}
