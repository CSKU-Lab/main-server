//go:build e2e
// +build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GradingRoutesTestSuite tests all grading-related endpoints
func TestGradingRoutes(t *testing.T) {
	suite := &GradingRoutesTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	t.Run("GetGradebook_AdminCanGet", suite.TestGetGradebook_AdminCanGet)
	t.Run("GetGradebook_InstructorCanGet", suite.TestGetGradebook_InstructorCanGet)
	t.Run("GetGradebook_StudentCannotGet", suite.TestGetGradebook_StudentCannotGet)
	t.Run("ExportGradebookCSV_AdminCanExport", suite.TestExportGradebookCSV_AdminCanExport)
	t.Run("ExportGradebookXLSX_AdminCanExport", suite.TestExportGradebookXLSX_AdminCanExport)
	t.Run("ExportGradebook_InvalidFormat", suite.TestExportGradebook_InvalidFormat)
	t.Run("GetLabStudentStatus_AdminCanGet", suite.TestGetLabStudentStatus_AdminCanGet)
	t.Run("GetStudentSubmissionsByMaterial_AdminCanGet", suite.TestGetStudentSubmissionsByMaterial_AdminCanGet)
}

type GradingRoutesTestSuite struct {
	TestSuite
}

// TestGetGradebook_AdminCanGet tests admin getting gradebook
func (s *GradingRoutesTestSuite) TestGetGradebook_AdminCanGet() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, lab, and material
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Create material
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/gradebook", nil)
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
	// Gradebook structure varies, but should contain data
	assert.NotNil(s.T(), result)
}

// TestGetGradebook_InstructorCanGet tests instructor getting gradebook
func (s *GradingRoutesTestSuite) TestGetGradebook_InstructorCanGet() {
	// Create instructor and student users
	instructorID := s.CreateTestUser("instructor", []string{"instructor"})
	defer s.CleanupTestUser(instructorID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, lab, and material
	courseID := s.CreateTestCourse(instructorID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{instructorID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, instructorID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Create material
	materialID := s.CreateTestMaterial(labID, "code", instructorID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate instructor token
	instructorToken := s.GenerateTestJWT(instructorID, "test_instructor", []string{"instructor"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/gradebook", nil)
	req.Header.Set("Cookie", "access_token="+instructorToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
}

// TestGetGradebook_StudentCannotGet tests student trying to get gradebook
func (s *GradingRoutesTestSuite) TestGetGradebook_StudentCannotGet() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course and section
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/gradebook", nil)
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 403 Forbidden
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

// TestExportGradebookCSV_AdminCanExport tests admin exporting gradebook as CSV
func (s *GradingRoutesTestSuite) TestExportGradebookCSV_AdminCanExport() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, lab, and material
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Create material
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/gradebook/export?format=csv", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	assert.Equal(s.T(), "text/csv", contentType)

	// Verify content disposition
	contentDisposition := resp.Header.Get("Content-Disposition")
	assert.Contains(s.T(), contentDisposition, "gradebook.csv")
}

// TestExportGradebookXLSX_AdminCanExport tests admin exporting gradebook as XLSX
func (s *GradingRoutesTestSuite) TestExportGradebookXLSX_AdminCanExport() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, lab, and material
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Create material
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/gradebook/export?format=xlsx", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	assert.Equal(s.T(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", contentType)

	// Verify content disposition
	contentDisposition := resp.Header.Get("Content-Disposition")
	assert.Contains(s.T(), contentDisposition, "gradebook.xlsx")
}

// TestExportGradebook_InvalidFormat tests exporting with invalid format
func (s *GradingRoutesTestSuite) TestExportGradebook_InvalidFormat() {
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

	// Make request with invalid format
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/gradebook/export?format=pdf", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 400 Bad Request
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

// TestGetLabStudentStatus_AdminCanGet tests admin getting lab student status
func (s *GradingRoutesTestSuite) TestGetLabStudentStatus_AdminCanGet() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, lab, and material
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Create material
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/labs/"+labID+"/student-status", nil)
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
	// Result should contain student status data
	assert.NotNil(s.T(), result)
}

// TestGetStudentSubmissionsByMaterial_AdminCanGet tests admin getting student submissions by material
func (s *GradingRoutesTestSuite) TestGetStudentSubmissionsByMaterial_AdminCanGet() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, lab, and material
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{adminID}, []string{studentID})
	defer s.CleanupTestSection(sectionID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create lab section association
	labSectionID := s.CreateTestLabSection(labID, sectionID)
	_ = labSectionID

	// Create material
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request with student_id query param
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/labs/"+labID+"/materials/"+materialID+"/submissions?student_id="+studentID, nil)
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
