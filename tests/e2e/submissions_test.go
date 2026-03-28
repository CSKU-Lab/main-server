//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// SubmissionsTestSuite tests submission routes
// Routes tested:
// - POST /submissions - Create submission
// - GET /submissions/:id - Get submission details
// - GET /submissions/:id/listen - Listen to submission updates (SSE)
// - GET /submissions - List user submissions
// - PATCH /cms/submissions/:id/manual-score - Update manual score (admin/instructor)
type SubmissionsTestSuite struct {
	TestSuite
}

func TestSubmissionsTestSuite(t *testing.T) {
	CheckE2EEnabled(t)
	suite.Run(t, new(SubmissionsTestSuite))
}

// TestCreateSubmission_Student_Success tests student creating a submission
func (s *SubmissionsTestSuite) TestCreateSubmission_Student_Success() {
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

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"lab_id":      labID,
		"material_id": materialID,
		"section_id":  sectionID,
		"payload": map[string]interface{}{
			"code":     "print('Hello World')",
			"language": "python",
		},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/submissions"), reqBody, studentToken)

	s.AssertSuccess(resp)

	submissionID := s.ExtractIDFromResponse(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestCreateSubmission_Unauthorized tests creating submission without auth
func (s *SubmissionsTestSuite) TestCreateSubmission_Unauthorized() {
	reqBody := map[string]interface{}{
		"lab_id":      "some-lab-id",
		"material_id": "some-material-id",
		"section_id":  "some-section-id",
		"payload": map[string]interface{}{
			"code": "print('test')",
		},
	}

	resp := s.RequestWithoutAuth("POST", BuildURL("/submissions"), reqBody)

	s.AssertUnauthorized(resp)
}

// TestCreateSubmission_InvalidData tests creating submission with invalid data
func (s *SubmissionsTestSuite) TestCreateSubmission_InvalidData() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"lab_id":      "invalid-uuid",
		"material_id": "invalid-uuid",
		// Missing required fields
	}

	resp := s.RequestWithAuth("POST", BuildURL("/submissions"), reqBody, studentToken)

	s.AssertBadRequest(resp)
}

// TestCreateSubmission_MissingPayload tests creating submission without payload
func (s *SubmissionsTestSuite) TestCreateSubmission_MissingPayload() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"lab_id":      "some-lab-id",
		"material_id": "some-material-id",
		// Missing payload
	}

	resp := s.RequestWithAuth("POST", BuildURL("/submissions"), reqBody, studentToken)

	s.AssertBadRequest(resp)
}

// TestGetSubmission_Owner_Success tests student getting their own submission
func (s *SubmissionsTestSuite) TestGetSubmission_Owner_Success() {
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

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/submissions/"+submissionID), nil, studentToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().Equal(submissionID, result["id"])

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetSubmission_Instructor_Success tests instructor getting student's submission
func (s *SubmissionsTestSuite) TestGetSubmission_Instructor_Success() {
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

	resp := s.RequestWithAuth("GET", BuildURL("/submissions/"+submissionID), nil, instructorToken)

	s.AssertSuccess(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetSubmission_NotFound tests getting non-existent submission
func (s *SubmissionsTestSuite) TestGetSubmission_NotFound() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/submissions/nonexistent-submission-id"), nil, studentToken)

	s.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

// TestListSubmissions_Student_Success tests student listing their submissions
func (s *SubmissionsTestSuite) TestListSubmissions_Student_Success() {
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

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/submissions"), nil, studentToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestListSubmissions_WithFilters tests listing submissions with filters
func (s *SubmissionsTestSuite) TestListSubmissions_WithFilters() {
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

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/submissions?material_id="+materialID+"&lab_id="+labID+"&section_id="+sectionID), nil, studentToken)

	s.AssertSuccess(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestUpdateManualScore_Instructor_Success tests instructor updating manual score
func (s *SubmissionsTestSuite) TestUpdateManualScore_Instructor_Success() {
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
		"manual_score": 95.5,
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/submissions/"+submissionID+"/manual-score"), reqBody, instructorToken)

	s.AssertNoContent(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestUpdateManualScore_Admin_Success tests admin updating manual score
func (s *SubmissionsTestSuite) TestUpdateManualScore_Admin_Success() {
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

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"manual_score": 88.0,
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/submissions/"+submissionID+"/manual-score"), reqBody, adminToken)

	s.AssertNoContent(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestUpdateManualScore_Student_Forbidden tests student trying to update manual score
func (s *SubmissionsTestSuite) TestUpdateManualScore_Student_Forbidden() {
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

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"manual_score": 100.0,
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/submissions/"+submissionID+"/manual-score"), reqBody, studentToken)

	s.AssertForbidden(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestUpdateManualScore_InvalidScore tests updating manual score with invalid value
func (s *SubmissionsTestSuite) TestUpdateManualScore_InvalidScore() {
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
		"manual_score": "invalid",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/submissions/"+submissionID+"/manual-score"), reqBody, instructorToken)

	s.AssertBadRequest(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM code_submissions WHERE submission_id = $1", submissionID)
	s.DB.Exec("DELETE FROM submissions WHERE id = $1", submissionID)
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}
