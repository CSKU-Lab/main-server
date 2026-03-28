//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/infrastructure/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateTestUserWithUsername creates a test user and returns both ID and username
func (s *TestSuite) CreateTestUserWithUsername(role string, roles []string) (string, string) {
	userID := uuid.New().String()
	username := fmt.Sprintf("test_%s_%s", role, uuid.New().String()[:8])
	email := fmt.Sprintf("%s@e2etest.example.com", username)
	now := time.Now()

	// Hash password
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.DefaultCost)

	// Convert roles to PostgreSQL array format
	rolesArray := "{}"
	if len(roles) > 0 {
		rolesArray = "{"
		for i, r := range roles {
			if i > 0 {
				rolesArray += ","
			}
			rolesArray += r
		}
		rolesArray += "}"
	}

	// Insert user
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO users (id, email, type, username, display_name, roles, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6::role[], $7, $8, false)
	`, userID, email, "credential", username, fmt.Sprintf("Test %s", role), rolesArray, now, now)
	s.Require().NoError(err, "Failed to create test user")

	// Insert password
	_, err = s.DB.ExecContext(s.Ctx, `
		INSERT INTO user_passwords (user_id, password)
		VALUES ($1, $2)
	`, userID, string(passwordHash))
	s.Require().NoError(err, "Failed to create user password")

	return userID, username
}

// CreateTestUser creates a test user with the specified role and returns the user ID
func (s *TestSuite) CreateTestUser(role string, roles []string) string {
	userID, _ := s.CreateTestUserWithUsername(role, roles)
	return userID
}

// CreateTestCourse creates a test course and returns the course ID
func (s *TestSuite) CreateTestCourse(creatorID string) string {
	courseID := uuid.New().String()
	now := time.Now()

	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO courses (id, name, visibility, created_at, updated_at, is_deleted, is_archived)
		VALUES ($1, $2, $3, $4, $5, false, false)
	`, courseID, fmt.Sprintf("E2E Test Course %s", uuid.New().String()[:8]), "public", now, now)
	s.Require().NoError(err, "Failed to create test course")

	// Add creator - schema uses creator_id and order, not user_id
	_, err = s.DB.ExecContext(s.Ctx, `
		INSERT INTO course_creators (course_id, creator_id, "order")
		VALUES ($1, $2, $3)
	`, courseID, creatorID, 1)
	s.Require().NoError(err, "Failed to add course creator")

	return courseID
}

// CreateTestSection creates a test section for a course and returns the section ID
func (s *TestSuite) CreateTestSection(courseID string, semesterID string, instructorIDs []string, studentIDs []string) string {
	sectionID := uuid.New().String()
	now := time.Now()

	if semesterID == "" {
		// Create a default semester if not provided
		semesterID = s.CreateTestSemester()
	}

	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO sections (id, name, course_id, semester_id, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, false)
	`, sectionID, fmt.Sprintf("E2E Test Section %s", uuid.New().String()[:8]), courseID, semesterID, now, now)
	s.Require().NoError(err, "Failed to create test section")

	// Add instructors - schema uses composite PK (section_id, instructor_id), no id or timestamps
	for _, instructorID := range instructorIDs {
		_, err = s.DB.ExecContext(s.Ctx, `
			INSERT INTO section_instructors (section_id, instructor_id)
			VALUES ($1, $2)
		`, sectionID, instructorID)
		s.Require().NoError(err, "Failed to add section instructor")
	}

	// Add students - schema uses composite PK (section_id, student_id) with is_deleted flag
	for _, studentID := range studentIDs {
		_, err = s.DB.ExecContext(s.Ctx, `
			INSERT INTO section_students (section_id, student_id, is_deleted)
			VALUES ($1, $2, false)
		`, sectionID, studentID)
		s.Require().NoError(err, "Failed to add section student")
	}

	return sectionID
}

// CreateTestSemester creates a test semester and returns the semester ID
func (s *TestSuite) CreateTestSemester() string {
	semesterID := uuid.New().String()
	now := time.Now()

	// Schema uses: id, name, type, started_date (not start_date), created_at, updated_at, is_deleted, deleted_at
	// Note: No end_date field in schema
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO semesters (id, name, type, started_date, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, false)
	`, semesterID, fmt.Sprintf("E2E Test Semester %s", uuid.New().String()[:8]), "first", now, now, now)
	s.Require().NoError(err, "Failed to create test semester")

	return semesterID
}

// CreateTestLab creates a test lab for a course and returns the lab ID
func (s *TestSuite) CreateTestLab(courseID string, createdBy string) string {
	labID := uuid.New().String()
	now := time.Now()

	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO labs (id, display_name, course_id, created_by, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, false)
	`, labID, fmt.Sprintf("E2E Test Lab %s", uuid.New().String()[:8]), courseID, createdBy, now, now)
	s.Require().NoError(err, "Failed to create test lab")

	return labID
}

// CreateTestLabSection creates a lab section association
func (s *TestSuite) CreateTestLabSection(labID string, sectionID string) string {
	labSectionID := uuid.New().String()
	now := time.Now()

	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO lab_sections (id, lab_id, section_id, position, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, labSectionID, labID, sectionID, 1, "hidden", now, now)
	s.Require().NoError(err, "Failed to create test lab section")

	return labSectionID
}

// CreateTestMaterial creates a test material for a lab and returns the material ID
func (s *TestSuite) CreateTestMaterial(labID string, materialType string, createdBy string) string {
	materialID := uuid.New().String()
	now := time.Now()

	// Schema requires: id, name, type, visibility, created_by, auto_score, manual_score, created_at, updated_at, is_deleted
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO materials (id, name, type, visibility, created_by, auto_score, manual_score, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false)
	`, materialID, fmt.Sprintf("E2E Test Material %s", uuid.New().String()[:8]), materialType, "public", createdBy, 0, 0, now, now)
	s.Require().NoError(err, "Failed to create test material")

	// Create lab material association - schema doesn't have position column
	labMaterialID := uuid.New().String()
	_, err = s.DB.ExecContext(s.Ctx, `
		INSERT INTO lab_materials (id, lab_id, material_id, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, false)
	`, labMaterialID, labID, materialID, now, now)
	s.Require().NoError(err, "Failed to create test lab material")

	return materialID
}

// CreateTestMaterialStandalone creates a test material without lab association
func (s *TestSuite) CreateTestMaterialStandalone(materialType string, createdBy string) string {
	materialID := uuid.New().String()
	now := time.Now()

	// Schema requires: id, name, type, visibility, created_by, auto_score, manual_score, created_at, updated_at, is_deleted
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO materials (id, name, type, visibility, created_by, auto_score, manual_score, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false)
	`, materialID, fmt.Sprintf("E2E Test Material %s", uuid.New().String()[:8]), materialType, "public", createdBy, 0, 0, now, now)
	s.Require().NoError(err, "Failed to create test material")

	return materialID
}
func (s *TestSuite) CreateTestCodeMaterial(labID string, taskID string, createdBy string) string {
	materialID := s.CreateTestMaterial(labID, "code", createdBy)

	// Schema: material_id, description, task_id, hide_test_cases
	// No id, time_limit, memory_limit, created_at, updated_at columns
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO code_materials (material_id, description, task_id, hide_test_cases)
		VALUES ($1, $2, $3, $4)
	`, materialID, "Test code material description", taskID, true)
	s.Require().NoError(err, "Failed to create test code material")

	return materialID
}

// CreateTestSubmission creates a test submission and returns the submission ID
func (s *TestSuite) CreateTestSubmission(userID string, materialID string, labID string, sectionID string, courseID string) string {
	submissionID := uuid.New().String()
	now := time.Now()

	// Schema requires: id, user_id, material_id, lab_id, section_id, course_id, created_at, updated_at, status, submission_order, auto_score, manual_score, ip_address
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO submissions (id, user_id, material_id, lab_id, section_id, course_id, created_at, updated_at, status, submission_order, auto_score, manual_score, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, submissionID, userID, materialID, labID, sectionID, courseID, now, now, "queued", 1, 0, 0, "127.0.0.1")
	s.Require().NoError(err, "Failed to create test submission")

	// Create code submission - schema uses files (jsonb), status, avg_wall_time, avg_memory, test_case_groups
	// No id, code, language, created_at, updated_at columns
	filesJSON := `[{"name": "main.py", "content": "print('Hello World')"}]`
	_, err = s.DB.ExecContext(s.Ctx, `
		INSERT INTO code_submissions (submission_id, files, status)
		VALUES ($1, $2::jsonb, $3)
	`, submissionID, filesJSON, "pending")
	s.Require().NoError(err, "Failed to create test code submission")

	return submissionID
}

// GenerateTestJWT generates a JWT token for a test user
func (s *TestSuite) GenerateTestJWT(userID string, username string, roles []string) string {
	user := &models.User{
		ID:          userID,
		Username:    username,
		DisplayName: fmt.Sprintf("Test %s", username),
		Roles:       s.convertRoles(roles),
	}

	token, err := auth.SignAccessToken(user, s.GetJWTSecret())
	s.Require().NoError(err, "Failed to generate test JWT")

	return token
}

// GenerateTestRefreshToken generates a refresh token for a test user
func (s *TestSuite) GenerateTestRefreshToken(userID string) string {
	token, err := auth.SignRefreshToken(userID, s.GetRefreshSecret())
	s.Require().NoError(err, "Failed to generate test refresh token")

	return token
}

// convertRoles converts string roles to models.Role slice
func (s *TestSuite) convertRoles(roles []string) []models.Role {
	result := make([]models.Role, len(roles))
	for i, r := range roles {
		result[i] = models.Role(r)
	}
	return result
}

// StoreRefreshToken stores a refresh token in the database
func (s *TestSuite) StoreRefreshToken(userID string, token string) {
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO user_refresh_tokens (user_id, token)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET token = $2
	`, userID, token)
	s.Require().NoError(err, "Failed to store refresh token")
}

// CleanupTestUser removes a test user and all related data
func (s *TestSuite) CleanupTestUser(userID string) {
	// Delete related data first
	s.DB.Exec("DELETE FROM user_passwords WHERE user_id = $1", userID)
	s.DB.Exec("DELETE FROM user_refresh_tokens WHERE user_id = $1", userID)
	s.DB.Exec("DELETE FROM section_students WHERE student_id = $1", userID)
	s.DB.Exec("DELETE FROM section_instructors WHERE instructor_id = $1", userID)
	s.DB.Exec("DELETE FROM course_creators WHERE creator_id = $1", userID)

	// Delete submissions
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id IN (SELECT id FROM submissions WHERE user_id = $1)", userID)
	s.DB.Exec("DELETE FROM submissions WHERE user_id = $1", userID)

	// Delete user
	s.DB.Exec("DELETE FROM users WHERE id = $1", userID)
}

// CleanupTestCourse removes a test course and all related data
func (s *TestSuite) CleanupTestCourse(courseID string) {
	// Get all sections for this course
	var sectionIDs []string
	err := s.DB.Select(&sectionIDs, "SELECT id FROM sections WHERE course_id = $1", courseID)
	if err == nil {
		for _, sectionID := range sectionIDs {
			s.CleanupTestSection(sectionID)
		}
	}

	// Get all labs for this course
	var labIDs []string
	err = s.DB.Select(&labIDs, "SELECT id FROM labs WHERE course_id = $1", courseID)
	if err == nil {
		for _, labID := range labIDs {
			s.CleanupTestLab(labID)
		}
	}

	// Delete course creators
	s.DB.Exec("DELETE FROM course_creators WHERE course_id = $1", courseID)

	// Delete course
	s.DB.Exec("DELETE FROM courses WHERE id = $1", courseID)
}

// CleanupTestSection removes a test section and all related data
func (s *TestSuite) CleanupTestSection(sectionID string) {
	// Delete lab sections
	s.DB.Exec("DELETE FROM lab_sections WHERE section_id = $1", sectionID)

	// Delete section students
	s.DB.Exec("DELETE FROM section_students WHERE section_id = $1", sectionID)

	// Delete section instructors
	s.DB.Exec("DELETE FROM section_instructors WHERE section_id = $1", sectionID)

	// Delete section logs
	s.DB.Exec("DELETE FROM section_logs WHERE section_id = $1", sectionID)

	// Delete section
	s.DB.Exec("DELETE FROM sections WHERE id = $1", sectionID)
}

// CleanupTestLab removes a test lab and all related data
func (s *TestSuite) CleanupTestLab(labID string) {
	// Delete lab materials
	s.DB.Exec("DELETE FROM lab_materials WHERE lab_id = $1", labID)

	// Delete lab sections
	s.DB.Exec("DELETE FROM lab_sections WHERE lab_id = $1", labID)

	// Delete default labs
	s.DB.Exec("DELETE FROM default_labs WHERE lab_id = $1", labID)

	// Delete code materials for this lab
	var materialIDs []string
	err := s.DB.Select(&materialIDs, "SELECT material_id FROM lab_materials WHERE lab_id = $1", labID)
	if err == nil {
		for _, materialID := range materialIDs {
			s.DB.Exec("DELETE FROM code_materials WHERE material_id = $1", materialID)
			s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
		}
	}

	// Delete lab
	s.DB.Exec("DELETE FROM labs WHERE id = $1", labID)
}

// Context returns the test context
func (s *TestSuite) Context() context.Context {
	return s.Ctx
}
