//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SectionsRoutesTestSuite tests all section management endpoints
type SectionsRoutesTestSuite struct {
	TestSuite
}

func TestSectionsRoutes(t *testing.T) {
	suite.Run(t, new(SectionsRoutesTestSuite))
}

// TestListSections_AdminCanList tests admin listing sections
func (s *SectionsRoutesTestSuite) TestListSections_AdminCanList() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and section
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections?page=1&page_size=10", nil)
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

// TestCreateSection_AdminCanCreate tests admin creating a section
func (s *SectionsRoutesTestSuite) TestCreateSection_AdminCanCreate() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and semester
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "E2E Test Section")
	writer.WriteField("semester_id", semesterID)
	writer.WriteField("course_id", courseID)
	writer.WriteField("instructors[]", adminID)
	writer.Close()

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/sections", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["id"])

	// Cleanup
	if sectionID, ok := result["id"].(string); ok {
		s.CleanupTestSection(sectionID)
	}
}

// TestCreateSection_InstructorCanCreate tests instructor creating a section
func (s *SectionsRoutesTestSuite) TestCreateSection_InstructorCanCreate() {
	// Create instructor user
	instructorID := s.CreateTestUser("instructor", []string{"instructor"})
	defer s.CleanupTestUser(instructorID)

	// Create a test course and semester
	courseID := s.CreateTestCourse(instructorID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()

	// Generate instructor token
	instructorToken := s.GenerateTestJWT(instructorID, "test_instructor", []string{"instructor"})

	// Prepare multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "E2E Test Section by Instructor")
	writer.WriteField("semester_id", semesterID)
	writer.WriteField("course_id", courseID)
	writer.WriteField("instructors[]", instructorID)
	writer.Close()

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/sections", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "access_token="+instructorToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["id"])

	// Cleanup
	if sectionID, ok := result["id"].(string); ok {
		s.CleanupTestSection(sectionID)
	}
}

// TestCreateSection_StudentCannotCreate tests student trying to create section
func (s *SectionsRoutesTestSuite) TestCreateSection_StudentCannotCreate() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course and semester
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Prepare multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "E2E Test Section by Student")
	writer.WriteField("semester_id", semesterID)
	writer.WriteField("course_id", courseID)
	writer.Close()

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/sections", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 403 Forbidden
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

// TestGetSectionByID_Success tests getting a section by ID
func (s *SectionsRoutesTestSuite) TestGetSectionByID_Success() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and section
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID, nil)
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
	assert.Equal(s.T(), sectionID, result["id"])
}

// TestUpdateSection_AdminCanUpdate tests admin updating a section
func (s *SectionsRoutesTestSuite) TestUpdateSection_AdminCanUpdate() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and section
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("name", "Updated Section Name")
	writer.WriteField("semester_id", semesterID)
	writer.Close()

	// Make request
	req := httptest.NewRequest("PATCH", "/api/v1/cms/sections/"+sectionID, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusAccepted, resp.StatusCode)
}

// TestDeleteSection_AdminCanDelete tests admin deleting a section
func (s *SectionsRoutesTestSuite) TestDeleteSection_AdminCanDelete() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and section
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	// Don't defer cleanup - we're testing deletion

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("DELETE", "/api/v1/cms/sections/"+sectionID, nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
}

// TestGetSectionStudents_AdminCanGet tests admin getting section students
func (s *SectionsRoutesTestSuite) TestGetSectionStudents_AdminCanGet() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course and section with student
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/students", nil)
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
	assert.NotNil(s.T(), result["data"])
}

// TestAddStudentsToSection_AdminCanAdd tests admin adding students to section
func (s *SectionsRoutesTestSuite) TestAddStudentsToSection_AdminCanAdd() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Get the student's username from database
	var studentUsername string
	err := s.DB.Get(&studentUsername, "SELECT username FROM users WHERE id = $1", studentID)
	s.Require().NoError(err, "Failed to get student username")

	// Create a test course and section
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare request with actual username
	addReq := map[string]interface{}{
		"student_usernames": []string{studentUsername},
	}
	reqBody, _ := json.Marshal(addReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/sections/"+sectionID+"/students", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
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
	assert.Equal(s.T(), "Students added successfully", result["message"])
}

// TestRemoveStudentsFromSection_AdminCanRemove tests admin removing students from section
func (s *SectionsRoutesTestSuite) TestRemoveStudentsFromSection_AdminCanRemove() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course and section with student
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare request
	removeReq := map[string]interface{}{
		"student_ids": []string{studentID},
	}
	reqBody, _ := json.Marshal(removeReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/sections/"+sectionID+"/students/remove", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
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
	assert.Equal(s.T(), "Students removed successfully", result["message"])
}

// TestGetSectionLabs_AdminCanGet tests admin getting section labs
func (s *SectionsRoutesTestSuite) TestGetSectionLabs_AdminCanGet() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course, section, and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID // We don't have a cleanup function for this yet

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/labs", nil)
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

// TestAddLabToSection_AdminCanAdd tests admin adding lab to section
func (s *SectionsRoutesTestSuite) TestAddLabToSection_AdminCanAdd() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course, section, and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare request - SetLabSection expects lab_ids array
	addReq := map[string]interface{}{
		"lab_ids": []string{labID},
	}
	reqBody, _ := json.Marshal(addReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/sections/"+sectionID+"/labs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
}

// TestUpdateLabSectionStatus_AdminCanUpdate tests admin updating lab section status
func (s *SectionsRoutesTestSuite) TestUpdateLabSectionStatus_AdminCanUpdate() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course, section, and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare request
	updateReq := map[string]interface{}{
		"status": "open",
	}
	reqBody, _ := json.Marshal(updateReq)

	// Make request
	req := httptest.NewRequest("PATCH", "/api/v1/cms/sections/"+sectionID+"/labs/"+labID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusAccepted, resp.StatusCode)
}

// TestGetLabInSection_AdminCanGet tests admin getting lab details in section
func (s *SectionsRoutesTestSuite) TestGetLabInSection_AdminCanGet() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course, section, and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/labs/"+labID, nil)
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
	assert.NotNil(s.T(), result["lab_name"])
	assert.NotNil(s.T(), result["status"])
}

// TestGetSectionLogs_AdminCanGet tests admin getting section logs
func (s *SectionsRoutesTestSuite) TestGetSectionLogs_AdminCanGet() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and section
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{})
	defer s.CleanupTestSection(sectionID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/logs", nil)
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

// TestStudentListSections_StudentCanListOwn tests student listing their own sections
func (s *SectionsRoutesTestSuite) TestStudentListSections_StudentCanListOwn() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course and section with student
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/sections?page=1&page_size=10", nil)
	req.Header.Set("Cookie", "access_token="+studentToken)

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

// TestStudentGetSection_StudentCanGetOwn tests student getting their own section
func (s *SectionsRoutesTestSuite) TestStudentGetSection_StudentCanGetOwn() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course and section with student
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/sections/"+sectionID, nil)
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result["section"])
	assert.NotNil(s.T(), result["course"])
}

// TestStudentUnenroll_StudentCanUnenroll tests student unenrolling from section
func (s *SectionsRoutesTestSuite) TestStudentUnenroll_StudentCanUnenroll() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course and section with student
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("DELETE", "/api/v1/sections/"+sectionID+"/unenroll", nil)
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Students removed successfully", result["message"])
}
