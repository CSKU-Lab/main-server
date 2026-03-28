package integration_test

import (
	"context"
	"testing"

	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLabUniquenessConstraint verifies that lab names are unique per course,
// not globally unique. This tests the fix for GitHub issue #13.
func TestLabUniquenessConstraint(t *testing.T) {
	// Skip if no database connection is available
	// This test requires a running PostgreSQL database
	db, err := setupTestDatabase(t)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	labRepo := sqlxAdapter.NewSqlxLabRepository(db)

	// Create test data
	course1ID := uuid.New().String()
	course2ID := uuid.New().String()
	userID := uuid.New().String()
	labName := "Introduction to Go"

	// Setup: Create test courses and user
	setupTestData(t, db, course1ID, course2ID, userID)

	t.Run("same lab name can exist in different courses", func(t *testing.T) {
		// Create lab in course 1
		lab1ID := uuid.New().String()
		req1 := &requests.CreateLab{
			DisplayName: labName,
			CourseID:    course1ID,
		}
		err := labRepo.Create(ctx, lab1ID, req1, userID)
		require.NoError(t, err, "should create lab in course 1")

		// Create lab with same name in course 2 - should succeed
		lab2ID := uuid.New().String()
		req2 := &requests.CreateLab{
			DisplayName: labName,
			CourseID:    course2ID,
		}
		err = labRepo.Create(ctx, lab2ID, req2, userID)
		require.NoError(t, err, "should create lab with same name in different course")

		// Verify both labs exist
		lab1, err := labRepo.GetByID(ctx, lab1ID)
		require.NoError(t, err)
		assert.Equal(t, labName, lab1.DisplayName)
		assert.Equal(t, course1ID, lab1.CourseID)

		lab2, err := labRepo.GetByID(ctx, lab2ID)
		require.NoError(t, err)
		assert.Equal(t, labName, lab2.DisplayName)
		assert.Equal(t, course2ID, lab2.CourseID)
	})

	t.Run("same lab name cannot exist twice in same course", func(t *testing.T) {
		// Create first lab in course 1
		labName2 := "Advanced Go Patterns"
		lab3ID := uuid.New().String()
		req3 := &requests.CreateLab{
			DisplayName: labName2,
			CourseID:    course1ID,
		}
		err := labRepo.Create(ctx, lab3ID, req3, userID)
		require.NoError(t, err, "should create first lab")

		// Try to create second lab with same name in course 1 - should fail
		lab4ID := uuid.New().String()
		req4 := &requests.CreateLab{
			DisplayName: labName2,
			CourseID:    course1ID,
		}
		err = labRepo.Create(ctx, lab4ID, req4, userID)
		require.Error(t, err, "should fail to create lab with duplicate name in same course")
		assert.Contains(t, err.Error(), "unique", "error should mention unique constraint violation")
	})

	t.Run("soft deleted lab allows reuse of name in same course", func(t *testing.T) {
		// Create a lab
		labName3 := "Temporary Lab"
		lab5ID := uuid.New().String()
		req5 := &requests.CreateLab{
			DisplayName: labName3,
			CourseID:    course1ID,
		}
		err := labRepo.Create(ctx, lab5ID, req5, userID)
		require.NoError(t, err, "should create lab")

		// Soft delete the lab
		err = labRepo.DeleteByID(ctx, lab5ID)
		require.NoError(t, err, "should delete lab")

		// Now create a new lab with the same name - should succeed
		lab6ID := uuid.New().String()
		req6 := &requests.CreateLab{
			DisplayName: labName3,
			CourseID:    course1ID,
		}
		err = labRepo.Create(ctx, lab6ID, req6, userID)
		require.NoError(t, err, "should create lab with same name after soft delete")
	})
}

// setupTestDatabase creates a connection to the test database
func setupTestDatabase(t *testing.T) (*sqlx.DB, error) {
	// Try to connect to the local development database
	// In a real CI environment, this would use testcontainers
	dsn := "postgres://cs_pg_user:cs_pg_password@localhost:5432/main-server?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// setupTestData creates prerequisite data (courses, user) for tests
func setupTestData(t *testing.T, db *sqlx.DB, course1ID, course2ID, userID string) {
	ctx := context.Background()

	// Use unique username/email based on userID to avoid conflicts
	uniqueSuffix := userID[:8]
	email := "test_" + uniqueSuffix + "@example.com"
	username := "testuser_" + uniqueSuffix

	// Create user with proper PostgreSQL array syntax
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, type, username, display_name, roles, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, ARRAY['instructor']::role[], NOW(), NOW(), false)
		ON CONFLICT (id) DO NOTHING
	`, userID, email, "oauth", username, "Test User")
	require.NoError(t, err, "should create test user")

	// Create courses with unique names
	course1Name := "Course 1 " + uniqueSuffix
	course2Name := "Course 2 " + uniqueSuffix

	_, err = db.ExecContext(ctx, `
		INSERT INTO courses (id, name, visibility, created_at, updated_at, is_deleted, is_archived)
		VALUES ($1, $2, $3, NOW(), NOW(), false, false)
		ON CONFLICT (id) DO NOTHING
	`, course1ID, course1Name, "public")
	require.NoError(t, err, "should create course 1")

	_, err = db.ExecContext(ctx, `
		INSERT INTO courses (id, name, visibility, created_at, updated_at, is_deleted, is_archived)
		VALUES ($1, $2, $3, NOW(), NOW(), false, false)
		ON CONFLICT (id) DO NOTHING
	`, course2ID, course2Name, "public")
	require.NoError(t, err, "should create course 2")
}

// TestLabSchemaMigration verifies the migration was applied correctly
func TestLabSchemaMigration(t *testing.T) {
	db, err := setupTestDatabase(t)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("old global unique index should not exist", func(t *testing.T) {
		var count int
		err := db.GetContext(ctx, &count, `
			SELECT COUNT(*) FROM pg_indexes 
			WHERE indexname = 'unique_display_name' 
			AND tablename = 'labs'
		`)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "old global unique_display_name index should not exist")
	})

	t.Run("new composite unique index should exist", func(t *testing.T) {
		var count int
		err := db.GetContext(ctx, &count, `
			SELECT COUNT(*) FROM pg_indexes 
			WHERE indexname = 'unique_display_name_per_course' 
			AND tablename = 'labs'
		`)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "new composite unique_display_name_per_course index should exist")
	})

	t.Run("new index should have correct columns", func(t *testing.T) {
		var indexDef string
		err := db.GetContext(ctx, &indexDef, `
			SELECT indexdef FROM pg_indexes 
			WHERE indexname = 'unique_display_name_per_course' 
			AND tablename = 'labs'
		`)
		require.NoError(t, err)
		assert.Contains(t, indexDef, "course_id", "index should include course_id column")
		assert.Contains(t, indexDef, "display_name", "index should include display_name column")
		assert.Contains(t, indexDef, "UNIQUE", "index should be unique")
		assert.Contains(t, indexDef, "is_deleted = false", "index should have soft-delete condition")
	})
}
