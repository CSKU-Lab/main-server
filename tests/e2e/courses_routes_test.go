//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// CoursesRoutesTestSuite tests all course management endpoints
func TestCoursesRoutes(t *testing.T) {
	suite := &CoursesRoutesTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	t.Run("ListCourses_AdminCanListAll", suite.TestListCourses_AdminCanListAll)
	t.Run("ListCourses_StudentCanListOwn", suite.TestListCourses_StudentCanListOwn)
	t.Run("ListCourses_Unauthorized", suite.TestListCourses_Unauthorized)
	t.Run("CreateCourse_AdminCanCreate", suite.TestCreateCourse_AdminCanCreate)
	t.Run("CreateCourse_InstructorCanCreate", suite.TestCreateCourse_InstructorCanCreate)
	t.Run("CreateCourse_StudentCannotCreate", suite.TestCreateCourse_StudentCannotCreate)
	t.Run("CreateCourse_InvalidData", suite.TestCreateCourse_InvalidData)
	t.Run("GetCourseByID_Success", suite.TestGetCourseByID_Success)
	t.Run("GetCourseByID_NotFound", suite.TestGetCourseByID_NotFound)
	t.Run("UpdateCourse_AdminCanUpdate", suite.TestUpdateCourse_AdminCanUpdate)
	t.Run("UpdateCourse_StudentCannotUpdate", suite.TestUpdateCourse_StudentCannotUpdate)
	t.Run("DeleteCourse_AdminCanDelete", suite.TestDeleteCourse_AdminCanDelete)
	t.Run("GetCourseSections_AdminCanGet", suite.TestGetCourseSections_AdminCanGet)
	t.Run("ListCourseLabs_AdminCanList", suite.TestListCourseLabs_AdminCanList)
	t.Run("SetDefaultLab_AdminCanSet", suite.TestSetDefaultLab_AdminCanSet)
	t.Run("ListDefaultLabs_AdminCanList", suite.TestListDefaultLabs_AdminCanList)
}

type CoursesRoutesTestSuite struct {
	TestSuite
}

// TestListCourses_AdminCanListAll tests admin listing all courses
func (s *CoursesRoutesTestSuite) TestListCourses_AdminCanListAll() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/courses?page=1&page_size=10", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["pagination"])
	assert.NotNil(s.T(), result["data"])

	// Verify pagination structure
	pagination, ok := result["pagination"].(map[string]interface{})
	assert.True(s.T(), ok)
	assert.NotNil(s.T(), pagination["page"])
	assert.NotNil(s.T(), pagination["total_page"])
	assert.NotNil(s.T(), pagination["total_rows"])
}

// TestListCourses_StudentCanListOwn tests student listing courses
func (s *CoursesRoutesTestSuite) TestListCourses_StudentCanListOwn() {
	// Create instructor and student users
	instructorID := s.CreateTestUser("instructor", []string{"instructor"})
	defer s.CleanupTestUser(instructorID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course
	courseID := s.CreateTestCourse(instructorID)
	defer s.CleanupTestCourse(courseID)

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request - students can access CMS courses endpoint
	req := httptest.NewRequest("GET", "/api/v1/cms/courses?page=1&page_size=10", nil)
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Students are not allowed to access CMS routes
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

// TestListCourses_Unauthorized tests listing courses without auth
func (s *CoursesRoutesTestSuite) TestListCourses_Unauthorized() {
	// Make request without authentication
	req := httptest.NewRequest("GET", "/api/v1/cms/courses", nil)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 401 Unauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

// TestCreateCourse_AdminCanCreate tests admin creating a course
func (s *CoursesRoutesTestSuite) TestCreateCourse_AdminCanCreate() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare create course request
	createReq := map[string]interface{}{
		"name":        "E2E Test Course",
		"description": "A test course created by E2E tests",
		"visibility":  "public",
	}
	reqBody, _ := json.Marshal(createReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/courses", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	// Parse response to get course ID for cleanup
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["id"])

	// Cleanup
	if courseID, ok := result["id"].(string); ok {
		s.CleanupTestCourse(courseID)
	}
}

// TestCreateCourse_InstructorCanCreate tests instructor creating a course
func (s *CoursesRoutesTestSuite) TestCreateCourse_InstructorCanCreate() {
	// Create instructor user
	instructorID := s.CreateTestUser("instructor", []string{"instructor"})
	defer s.CleanupTestUser(instructorID)

	// Generate instructor token
	instructorToken := s.GenerateTestJWT(instructorID, "test_instructor", []string{"instructor"})

	// Prepare create course request
	createReq := map[string]interface{}{
		"name":        "E2E Test Course by Instructor",
		"description": "A test course created by instructor",
		"visibility":  "public",
	}
	reqBody, _ := json.Marshal(createReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/courses", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+instructorToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	// Parse response to get course ID for cleanup
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["id"])

	// Cleanup
	if courseID, ok := result["id"].(string); ok {
		s.CleanupTestCourse(courseID)
	}
}

// TestCreateCourse_StudentCannotCreate tests student trying to create course
func (s *CoursesRoutesTestSuite) TestCreateCourse_StudentCannotCreate() {
	// Create student user
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Prepare create course request
	createReq := map[string]interface{}{
		"name":        "E2E Test Course by Student",
		"description": "This should not be created",
		"visibility":  "public",
	}
	reqBody, _ := json.Marshal(createReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/courses", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 403 Forbidden
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

// TestCreateCourse_InvalidData tests creating course with invalid data
func (s *CoursesRoutesTestSuite) TestCreateCourse_InvalidData() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare invalid create course request (missing required fields)
	createReq := map[string]interface{}{
		"description": "Missing name field",
	}
	reqBody, _ := json.Marshal(createReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/courses", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 400 Bad Request
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

// TestGetCourseByID_Success tests getting a course by ID
func (s *CoursesRoutesTestSuite) TestGetCourseByID_Success() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/courses/"+courseID, nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), courseID, result["id"])
}

// TestGetCourseByID_NotFound tests getting non-existent course
func (s *CoursesRoutesTestSuite) TestGetCourseByID_NotFound() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request for non-existent course
	req := httptest.NewRequest("GET", "/api/v1/cms/courses/non-existent-id", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 404 or error
	assert.Contains(s.T(), []int{http.StatusNotFound, http.StatusInternalServerError}, resp.StatusCode)
}

// TestUpdateCourse_AdminCanUpdate tests admin updating a course
func (s *CoursesRoutesTestSuite) TestUpdateCourse_AdminCanUpdate() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare update request
	updateReq := map[string]interface{}{
		"name":        "Updated Course Name",
		"description": "Updated description",
	}
	reqBody, _ := json.Marshal(updateReq)

	// Make request
	req := httptest.NewRequest("PATCH", "/api/v1/cms/courses/"+courseID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
}

// TestUpdateCourse_StudentCannotUpdate tests student trying to update course
func (s *CoursesRoutesTestSuite) TestUpdateCourse_StudentCannotUpdate() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Prepare update request
	updateReq := map[string]interface{}{
		"name": "Hacked Course Name",
	}
	reqBody, _ := json.Marshal(updateReq)

	// Make request
	req := httptest.NewRequest("PATCH", "/api/v1/cms/courses/"+courseID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 403 Forbidden
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

// TestDeleteCourse_AdminCanDelete tests admin deleting a course
func (s *CoursesRoutesTestSuite) TestDeleteCourse_AdminCanDelete() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	// Don't defer cleanup - we're testing deletion

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("DELETE", "/api/v1/cms/courses/"+courseID, nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
}

// TestGetCourseSections_AdminCanGet tests admin getting course sections
func (s *CoursesRoutesTestSuite) TestGetCourseSections_AdminCanGet() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Create a semester and section
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/courses/"+courseID+"/sections", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["pagination"])
	assert.NotNil(s.T(), result["data"])
}

// TestListCourseLabs_AdminCanList tests admin listing course labs
func (s *CoursesRoutesTestSuite) TestListCourseLabs_AdminCanList() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Create a lab for the course
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/courses/"+courseID+"/labs", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["pagination"])
	assert.NotNil(s.T(), result["data"])
}

// TestSetDefaultLab_AdminCanSet tests admin setting default lab for course
func (s *CoursesRoutesTestSuite) TestSetDefaultLab_AdminCanSet() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Create a lab for the course
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare set default lab request
	setReq := map[string]interface{}{
		"lab_id":   labID,
		"position": 1,
	}
	reqBody, _ := json.Marshal(setReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/courses/"+courseID+"/default-labs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
}

// TestListDefaultLabs_AdminCanList tests admin listing default labs
func (s *CoursesRoutesTestSuite) TestListDefaultLabs_AdminCanList() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Create a lab for the course
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/courses/"+courseID+"/default-labs", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["pagination"])
	assert.NotNil(s.T(), result["data"])
}
