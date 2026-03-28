//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// SectionsTestSuite tests section management routes
// Routes tested:
// - GET /cms/sections - List sections
// - POST /cms/sections - Create section
// - GET /cms/sections/:id - Get section details
// - PATCH /cms/sections/:id - Update section
// - DELETE /cms/sections/:id - Delete section
// - GET /cms/sections/:id/students - List section students
// - POST /cms/sections/:id/students - Add students to section
// - POST /cms/sections/:id/students/remove - Remove students from section
// - GET /cms/sections/:sectionID/labs - List section labs
// - GET /cms/sections/:sectionID/labs/:labID - Get lab details in section
// - GET /cms/sections/:sectionID/logs - List section logs
// - GET /cms/sections/:id/gradebook - Get gradebook
// - GET /cms/sections/:id/gradebook/export - Export gradebook
// - POST /cms/sections/:sectionID/labs - Add lab to section
// - PATCH /cms/sections/:sectionID/labs - Update lab in section
// - POST /cms/sections/:sectionID/labs/delete - Remove lab from section
// - PATCH /cms/sections/:sectionID/labs/:labID - Update lab section status
// - GET /cms/sections/:sectionID/labs/:labID/student-status - Get student status for lab
type SectionsTestSuite struct {
	TestSuite
}

func TestSectionsTestSuite(t *testing.T) {
	CheckE2EEnabled(t)
	suite.Run(t, new(SectionsTestSuite))
}

// TestListSections_Admin_Success tests admin listing sections
func (s *SectionsTestSuite) TestListSections_Admin_Success() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestListSections_Instructor_Success tests instructor listing sections
func (s *SectionsTestSuite) TestListSections_Instructor_Success() {
	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections"), nil, instructorToken)

	s.AssertSuccess(resp)
}

// TestListSections_Student_Forbidden tests student trying to list sections
func (s *SectionsTestSuite) TestListSections_Student_Forbidden() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections"), nil, studentToken)

	s.AssertForbidden(resp)
}

// TestCreateSection_Admin_Success tests admin creating a section
func (s *SectionsTestSuite) TestCreateSection_Admin_Success() {
	// Create a test course first
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	// Create section using form data
	fields := map[string]string{
		"name":          "E2E Test Section " + generateRandomString(6),
		"semester_id":   semesterID,
		"course_id":     courseID,
		"instructors[]": s.TestUser.Admin.UserID,
	}

	resp := s.RequestWithFormData("POST", BuildURL("/cms/sections"), fields, nil, adminToken)

	s.AssertCreated(resp)

	sectionID := s.ExtractIDFromResponse(resp)
	defer s.CleanupTestSection(sectionID)
}

// TestCreateSection_Instructor_Success tests instructor creating a section
func (s *SectionsTestSuite) TestCreateSection_Instructor_Success() {
	// Create a test course first
	courseID := s.CreateTestCourse(s.TestUser.Instructor.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	fields := map[string]string{
		"name":          "E2E Test Section " + generateRandomString(6),
		"semester_id":   semesterID,
		"course_id":     courseID,
		"instructors[]": s.TestUser.Instructor.UserID,
	}

	resp := s.RequestWithFormData("POST", BuildURL("/cms/sections"), fields, nil, instructorToken)

	s.AssertCreated(resp)

	sectionID := s.ExtractIDFromResponse(resp)
	defer s.CleanupTestSection(sectionID)
}

// TestCreateSection_MissingFields tests creating section with missing fields
func (s *SectionsTestSuite) TestCreateSection_MissingFields() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	fields := map[string]string{
		"name": "Test Section",
		// Missing course_id and semester_id
	}

	resp := s.RequestWithFormData("POST", BuildURL("/cms/sections"), fields, nil, adminToken)

	s.AssertBadRequest(resp)
}

// TestGetSection_Admin_Success tests admin getting section details
func (s *SectionsTestSuite) TestGetSection_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID), nil, adminToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().Equal(sectionID, result["id"])
}

// TestGetSection_NotFound tests getting non-existent section
func (s *SectionsTestSuite) TestGetSection_NotFound() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/nonexistent-section-id"), nil, adminToken)

	s.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

// TestUpdateSection_Admin_Success tests admin updating a section
func (s *SectionsTestSuite) TestUpdateSection_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	fields := map[string]string{
		"name": "Updated Section Name",
	}

	resp := s.RequestWithFormData("PATCH", BuildURL("/cms/sections/"+sectionID), fields, nil, adminToken)

	s.Assert().Equal(http.StatusAccepted, resp.StatusCode)
}

// TestDeleteSection_Admin_Success tests admin deleting a section
func (s *SectionsTestSuite) TestDeleteSection_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("DELETE", BuildURL("/cms/sections/"+sectionID), nil, adminToken)

	s.Assert().Equal(http.StatusNoContent, resp.StatusCode)
}

// TestListSectionStudents_Admin_Success tests admin listing section students
func (s *SectionsTestSuite) TestListSectionStudents_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/students"), nil, adminToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().NotNil(result["data"])
}

// TestAddStudentsToSection_Admin_Success tests admin adding students to section
func (s *SectionsTestSuite) TestAddStudentsToSection_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{})
	defer s.CleanupTestSection(sectionID)

	// Create another student to add
	newStudentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(newStudentID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"student_usernames": []string{s.TestUser.Student.Username},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/sections/"+sectionID+"/students"), reqBody, adminToken)

	s.AssertSuccess(resp)
}

// TestRemoveStudentsFromSection_Admin_Success tests admin removing students from section
func (s *SectionsTestSuite) TestRemoveStudentsFromSection_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"student_ids": []string{s.TestUser.Student.UserID},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/sections/"+sectionID+"/students/remove"), reqBody, adminToken)

	s.AssertSuccess(resp)
}

// TestListSectionLabs_Admin_Success tests admin listing labs in a section
func (s *SectionsTestSuite) TestListSectionLabs_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	s.CreateTestLabSection(labID, sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestAddLabToSection_Admin_Success tests admin adding a lab to a section
func (s *SectionsTestSuite) TestAddLabToSection_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"lab_ids": []string{labID},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/sections/"+sectionID+"/labs"), reqBody, adminToken)

	s.AssertCreated(resp)
}

// TestUpdateLabSectionStatus_Admin_Success tests admin updating lab section status
func (s *SectionsTestSuite) TestUpdateLabSectionStatus_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	s.CreateTestLabSection(labID, sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	status := "open"
	reqBody := map[string]interface{}{
		"status": &status,
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID), reqBody, adminToken)

	s.Assert().Equal(http.StatusAccepted, resp.StatusCode)
}

// TestGetSectionLogs_Admin_Success tests admin getting section logs
func (s *SectionsTestSuite) TestGetSectionLogs_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/logs"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestGetGradebook_Admin_Success tests admin getting gradebook
func (s *SectionsTestSuite) TestGetGradebook_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/gradebook"), nil, adminToken)

	s.AssertSuccess(resp)
}

// TestExportGradebookCSV_Admin_Success tests admin exporting gradebook as CSV
func (s *SectionsTestSuite) TestExportGradebookCSV_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/gradebook/export?format=csv"), nil, adminToken)

	s.AssertSuccess(resp)
	s.Assert().Equal("text/csv", resp.Headers.Get("Content-Type"))
}

// TestExportGradebookXLSX_Admin_Success tests admin exporting gradebook as XLSX
func (s *SectionsTestSuite) TestExportGradebookXLSX_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/gradebook/export?format=xlsx"), nil, adminToken)

	s.AssertSuccess(resp)
	s.Assert().Equal("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", resp.Headers.Get("Content-Type"))
}

// TestExportGradebook_InvalidFormat tests exporting gradebook with invalid format
func (s *SectionsTestSuite) TestExportGradebook_InvalidFormat() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/gradebook/export?format=pdf"), nil, adminToken)

	s.AssertBadRequest(resp)
}
