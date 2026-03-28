//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// CoursesTestSuite tests course management routes
// Routes tested:
// - GET /cms/courses - List courses
// - POST /cms/courses - Create course
// - GET /cms/courses/:courseID - Get course details
// - PATCH /cms/courses/:courseID - Update course
// - DELETE /cms/courses/:courseID - Delete course
// - GET /cms/courses/:courseID/sections - List course sections
// - GET /cms/courses/:courseID/labs - List course labs
// - GET /cms/courses/:courseID/default-labs - List default labs
// - POST /cms/courses/:courseID/default-labs - Set default lab
// - PATCH /cms/courses/:courseID/default-labs - Update default lab
// - POST /cms/courses/:courseID/default-labs/delete - Delete default lab
type CoursesTestSuite struct {
	TestSuite
}

func TestCoursesTestSuite(t *testing.T) {
	CheckE2EEnabled(t)
	suite.Run(t, new(CoursesTestSuite))
}

// TestListCourses_Admin_Success tests admin listing courses
func (s *CoursesTestSuite) TestListCourses_Admin_Success() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestListCourses_Instructor_Success tests instructor listing courses
func (s *CoursesTestSuite) TestListCourses_Instructor_Success() {
	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses"), nil, instructorToken)

	s.AssertSuccess(resp)
}

// TestListCourses_Student_Forbidden tests student trying to list courses
func (s *CoursesTestSuite) TestListCourses_Student_Forbidden() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses"), nil, studentToken)

	s.AssertForbidden(resp)
}

// TestListCourses_Unauthorized tests listing courses without auth
func (s *CoursesTestSuite) TestListCourses_Unauthorized() {
	resp := s.RequestWithoutAuth("GET", BuildURL("/cms/courses"), nil)

	s.AssertUnauthorized(resp)
}

// TestListCourses_WithFilters tests listing courses with filters
func (s *CoursesTestSuite) TestListCourses_WithFilters() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses?show=active&search=test"), nil, adminToken)

	s.AssertSuccess(resp)
}

// TestCreateCourse_Admin_Success tests admin creating a course
func (s *CoursesTestSuite) TestCreateCourse_Admin_Success() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"name":       "E2E Test Course " + generateRandomString(6),
		"visibility": "public",
		"creators":   []string{s.TestUser.Admin.UserID},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/courses"), reqBody, adminToken)

	s.AssertCreated(resp)

	// Extract course ID for cleanup
	courseID := s.ExtractIDFromResponse(resp)
	defer s.CleanupTestCourse(courseID)
}

// TestCreateCourse_Instructor_Success tests instructor creating a course
func (s *CoursesTestSuite) TestCreateCourse_Instructor_Success() {
	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	reqBody := map[string]interface{}{
		"name":       "E2E Test Course " + generateRandomString(6),
		"visibility": "public",
		"creators":   []string{s.TestUser.Instructor.UserID},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/courses"), reqBody, instructorToken)

	s.AssertCreated(resp)

	courseID := s.ExtractIDFromResponse(resp)
	defer s.CleanupTestCourse(courseID)
}

// TestCreateCourse_Student_Forbidden tests student trying to create course
func (s *CoursesTestSuite) TestCreateCourse_Student_Forbidden() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"name":       "Unauthorized Course",
		"visibility": "public",
		"creators":   []string{s.TestUser.Student.UserID},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/courses"), reqBody, studentToken)

	s.AssertForbidden(resp)
}

// TestCreateCourse_InvalidData tests creating course with invalid data
func (s *CoursesTestSuite) TestCreateCourse_InvalidData() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"name":       "Test Course",
		"visibility": "invalid_visibility",
		"creators":   []string{s.TestUser.Admin.UserID},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/courses"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestCreateCourse_MissingFields tests creating course with missing required fields
func (s *CoursesTestSuite) TestCreateCourse_MissingFields() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"name": "Test Course",
		// Missing visibility and creators
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/courses"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestGetCourse_Admin_Success tests admin getting course details
func (s *CoursesTestSuite) TestGetCourse_Admin_Success() {
	// Create a test course
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses/"+courseID), nil, adminToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().Equal(courseID, result["id"])
}

// TestGetCourse_NotFound tests getting non-existent course
func (s *CoursesTestSuite) TestGetCourse_NotFound() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses/nonexistent-course-id"), nil, adminToken)

	s.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

// TestUpdateCourse_Admin_Success tests admin updating a course
func (s *CoursesTestSuite) TestUpdateCourse_Admin_Success() {
	// Create a test course
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"name": "Updated Course Name",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/courses/"+courseID), reqBody, adminToken)

	s.AssertNoContent(resp)
}

// TestUpdateCourse_Instructor_OwnCourse tests instructor updating their own course
func (s *CoursesTestSuite) TestUpdateCourse_Instructor_OwnCourse() {
	// Create a test course with instructor as creator
	courseID := s.CreateTestCourse(s.TestUser.Instructor.UserID)
	defer s.CleanupTestCourse(courseID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	reqBody := map[string]interface{}{
		"name": "Updated by Instructor",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/courses/"+courseID), reqBody, instructorToken)

	s.AssertNoContent(resp)
}

// TestUpdateCourse_InvalidData tests updating course with invalid data
func (s *CoursesTestSuite) TestUpdateCourse_InvalidData() {
	// Create a test course
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"visibility": "invalid",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/courses/"+courseID), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestDeleteCourse_Admin_Success tests admin deleting a course
func (s *CoursesTestSuite) TestDeleteCourse_Admin_Success() {
	// Create a test course
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("DELETE", BuildURL("/cms/courses/"+courseID), nil, adminToken)

	s.AssertNoContent(resp)

	// Verify course is soft deleted
	var isDeleted bool
	err := s.DB.Get(&isDeleted, "SELECT is_deleted FROM courses WHERE id = $1", courseID)
	s.Require().NoError(err)
	s.Assert().True(isDeleted)
}

// TestDeleteCourse_Instructor_Forbidden tests instructor trying to delete course
func (s *CoursesTestSuite) TestDeleteCourse_Instructor_Forbidden() {
	// Create a test course
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("DELETE", BuildURL("/cms/courses/"+courseID), nil, instructorToken)

	s.AssertForbidden(resp)
}

// TestListCourseSections_Admin_Success tests admin listing course sections
func (s *CoursesTestSuite) TestListCourseSections_Admin_Success() {
	// Create a test course with a section
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses/"+courseID+"/sections"), nil, adminToken)

	s.AssertSuccess(resp)
}

// TestListCourseLabs_Admin_Success tests admin listing course labs
func (s *CoursesTestSuite) TestListCourseLabs_Admin_Success() {
	// Create a test course with a lab
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses/"+courseID+"/labs"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestSetDefaultLab_Admin_Success tests admin setting a default lab
func (s *CoursesTestSuite) TestSetDefaultLab_Admin_Success() {
	// Create a test course with a lab
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"lab_id":   labID,
		"position": 1,
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/courses/"+courseID+"/default-labs"), reqBody, adminToken)

	s.AssertCreated(resp)
}

// TestListDefaultLabs_Admin_Success tests admin listing default labs
func (s *CoursesTestSuite) TestListDefaultLabs_Admin_Success() {
	// Create a test course with a lab and default lab
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	// Set as default lab
	_, err := s.DB.ExecContext(s.Ctx, `
		INSERT INTO default_labs (id, course_id, lab_id, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, generateRandomString(32), courseID, labID, 1)
	s.Require().NoError(err)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/courses/"+courseID+"/default-labs"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}
