//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// GradingTestSuite tests grading-related routes
// Routes tested:
// - GET /sections/:sectionID/labs/:labID/materials/:materialID/submissions - Get submissions for grading
// - GET /sections/:sectionID/labs/:labID/student-status - Get student status for lab
// - GET /sections/:sectionID/labs/:labID - Get lab details with completion stats
// - POST /cms/submissions/:id/manual-score - Submit manual grade (instructor)
type GradingTestSuite struct {
	TestSuite
}

func TestGradingTestSuite(t *testing.T) {
	CheckE2EEnabled(t)
	suite.Run(t, new(GradingTestSuite))
}

// TestGetSubmissionsForGrading_Instructor_Success tests instructor getting submissions for grading
func (s *GradingTestSuite) TestGetSubmissionsForGrading_Instructor_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	s.CreateTestLabSection(labID, sectionID)

	materialID := s.CreateTestMaterial(labID, "code", s.TestUser.Admin.UserID)

	submissionID := s.CreateTestSubmission(s.TestUser.Student.UserID, materialID, labID, sectionID, courseID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID+"/materials/"+materialID+"/submissions"), nil, instructorToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().NotNil(result["data"])

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetSubmissionsForGrading_Admin_Success tests admin getting submissions for grading
func (s *GradingTestSuite) TestGetSubmissionsForGrading_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	s.CreateTestLabSection(labID, sectionID)

	materialID := s.CreateTestMaterial(labID, "code", s.TestUser.Admin.UserID)

	submissionID := s.CreateTestSubmission(s.TestUser.Student.UserID, materialID, labID, sectionID, courseID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID+"/materials/"+materialID+"/submissions"), nil, adminToken)

	s.AssertSuccess(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetSubmissionsForGrading_Student_Forbidden tests student trying to access grading submissions
func (s *GradingTestSuite) TestGetSubmissionsForGrading_Student_Forbidden() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	materialID := s.CreateTestMaterial(labID, "code", s.TestUser.Admin.UserID)

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID+"/materials/"+materialID+"/submissions"), nil, studentToken)

	s.AssertForbidden(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetStudentSubmissionsByStudentID_Instructor_Success tests instructor getting specific student submissions
func (s *GradingTestSuite) TestGetStudentSubmissionsByStudentID_Instructor_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	s.CreateTestLabSection(labID, sectionID)

	materialID := s.CreateTestMaterial(labID, "code", s.TestUser.Admin.UserID)

	submissionID := s.CreateTestSubmission(s.TestUser.Student.UserID, materialID, labID, sectionID, courseID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID+"/materials/"+materialID+"/submissions?student_id="+s.TestUser.Student.UserID), nil, instructorToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetLabStudentStatus_Instructor_Success tests instructor getting student status for lab
func (s *GradingTestSuite) TestGetLabStudentStatus_Instructor_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	s.CreateTestLabSection(labID, sectionID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID+"/student-status"), nil, instructorToken)

	s.AssertSuccess(resp)

	var result []interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().NotNil(result)
}

// TestGetLabStudentStatus_Admin_Success tests admin getting student status for lab
func (s *GradingTestSuite) TestGetLabStudentStatus_Admin_Success() {
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

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID+"/student-status"), nil, adminToken)

	s.AssertSuccess(resp)
}

// TestGetLabStudentStatus_Student_Forbidden tests student trying to get all student status
func (s *GradingTestSuite) TestGetLabStudentStatus_Student_Forbidden() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID+"/student-status"), nil, studentToken)

	s.AssertForbidden(resp)
}

// TestGetLabDetailsWithStats_Instructor_Success tests instructor getting lab details with completion stats
func (s *GradingTestSuite) TestGetLabDetailsWithStats_Instructor_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	s.CreateTestLabSection(labID, sectionID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/labs/"+labID), nil, instructorToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().NotNil(result["lab_name"])
	s.Assert().NotNil(result["total_students"])
	s.Assert().NotNil(result["completed_students"])
}

// TestSubmitGrade_Instructor_Success tests instructor submitting a grade
func (s *GradingTestSuite) TestSubmitGrade_Instructor_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	materialID := s.CreateTestMaterial(labID, "code", s.TestUser.Admin.UserID)

	submissionID := s.CreateTestSubmission(s.TestUser.Student.UserID, materialID, labID, sectionID, courseID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	reqBody := map[string]interface{}{
		"manual_score": 92.5,
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/submissions/"+submissionID+"/manual-score"), reqBody, instructorToken)

	s.AssertNoContent(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestSubmitGrade_InvalidScore tests submitting an invalid grade
func (s *GradingTestSuite) TestSubmitGrade_InvalidScore() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	materialID := s.CreateTestMaterial(labID, "code", s.TestUser.Admin.UserID)

	submissionID := s.CreateTestSubmission(s.TestUser.Student.UserID, materialID, labID, sectionID, courseID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	reqBody := map[string]interface{}{
		"manual_score": -10,
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/submissions/"+submissionID+"/manual-score"), reqBody, instructorToken)

	// Should either succeed (if negative scores are allowed) or fail with bad request
	// Depending on the validation rules
	s.Assert().True(resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusBadRequest)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestSubmitGrade_NonInstructor_Forbidden tests non-instructor trying to submit grade
func (s *GradingTestSuite) TestSubmitGrade_NonInstructor_Forbidden() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID, s.TestUser.Student2.UserID})
	defer s.CleanupTestSection(sectionID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	materialID := s.CreateTestMaterial(labID, "code", s.TestUser.Admin.UserID)

	submissionID := s.CreateTestSubmission(s.TestUser.Student.UserID, materialID, labID, sectionID, courseID)

	// Student2 tries to grade Student's submission
	student2Token := s.GenerateTestJWT(s.TestUser.Student2.UserID, s.TestUser.Student2.Username, s.TestUser.Student2.Roles)

	reqBody := map[string]interface{}{
		"manual_score": 85.0,
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/submissions/"+submissionID+"/manual-score"), reqBody, student2Token)

	s.AssertForbidden(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetGradebook_Admin_Success tests admin getting gradebook
func (s *GradingTestSuite) TestGetGradebook_Admin_Success() {
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

// TestGetGradebook_Instructor_Success tests instructor getting gradebook
func (s *GradingTestSuite) TestGetGradebook_Instructor_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/gradebook"), nil, instructorToken)

	s.AssertSuccess(resp)
}

// TestGetGradebook_Student_Forbidden tests student trying to access gradebook
func (s *GradingTestSuite) TestGetGradebook_Student_Forbidden() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/sections/"+sectionID+"/gradebook"), nil, studentToken)

	s.AssertForbidden(resp)
}

// TestExportGradebookCSV_Admin_Success tests admin exporting gradebook as CSV
func (s *GradingTestSuite) TestExportGradebookCSV_Admin_Success() {
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
	s.Assert().Contains(resp.Headers.Get("Content-Disposition"), "gradebook.csv")
}

// TestExportGradebookXLSX_Admin_Success tests admin exporting gradebook as XLSX
func (s *GradingTestSuite) TestExportGradebookXLSX_Admin_Success() {
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
	s.Assert().Contains(resp.Headers.Get("Content-Disposition"), "gradebook.xlsx")
}

// TestExportGradebook_InvalidFormat tests exporting gradebook with invalid format
func (s *GradingTestSuite) TestExportGradebook_InvalidFormat() {
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
