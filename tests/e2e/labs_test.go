//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// LabsTestSuite tests lab management routes
// Routes tested:
// - GET /cms/labs - List labs
// - POST /cms/labs - Create lab
// - GET /cms/labs/:labID - Get lab details
// - PATCH /cms/labs/:labID - Update lab
// - DELETE /cms/labs/:labID - Delete lab
// - GET /cms/labs/:labID/sections - List lab sections
// - GET /cms/labs/:labID/materials - List lab materials
// - GET /cms/labs/:labID/materials/all - Get all lab materials
// - POST /cms/labs/:labID/materials - Add material to lab
// - POST /cms/labs/:labID/materials/delete - Remove material from lab
type LabsTestSuite struct {
	TestSuite
}

func TestLabsTestSuite(t *testing.T) {
	CheckE2EEnabled(t)
	suite.Run(t, new(LabsTestSuite))
}

// TestListLabs_Admin_Success tests admin listing labs
func (s *LabsTestSuite) TestListLabs_Admin_Success() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestListLabs_Instructor_Success tests instructor listing labs
func (s *LabsTestSuite) TestListLabs_Instructor_Success() {
	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs"), nil, instructorToken)

	s.AssertSuccess(resp)
}

// TestListLabs_Student_Forbidden tests student trying to list labs
func (s *LabsTestSuite) TestListLabs_Student_Forbidden() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs"), nil, studentToken)

	s.AssertForbidden(resp)
}

// TestListLabs_WithFilters tests listing labs with filters
func (s *LabsTestSuite) TestListLabs_WithFilters() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs?course_id__is=some-course-id"), nil, adminToken)

	s.AssertSuccess(resp)
}

// TestCreateLab_Admin_Success tests admin creating a lab
func (s *LabsTestSuite) TestCreateLab_Admin_Success() {
	// Create a test course first
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"display_name": "E2E Test Lab " + generateRandomString(6),
		"course_id":    courseID,
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/labs"), reqBody, adminToken)

	s.AssertCreated(resp)

	labID := s.ExtractIDFromResponse(resp)
	defer s.CleanupTestLab(labID)
}

// TestCreateLab_Instructor_Success tests instructor creating a lab
func (s *LabsTestSuite) TestCreateLab_Instructor_Success() {
	// Create a test course first
	courseID := s.CreateTestCourse(s.TestUser.Instructor.UserID)
	defer s.CleanupTestCourse(courseID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	reqBody := map[string]interface{}{
		"display_name": "E2E Test Lab " + generateRandomString(6),
		"course_id":    courseID,
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/labs"), reqBody, instructorToken)

	s.AssertCreated(resp)

	labID := s.ExtractIDFromResponse(resp)
	defer s.CleanupTestLab(labID)
}

// TestCreateLab_Student_Forbidden tests student trying to create lab
func (s *LabsTestSuite) TestCreateLab_Student_Forbidden() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Unauthorized Lab",
		"course_id":    "some-course-id",
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/labs"), reqBody, studentToken)

	s.AssertForbidden(resp)
}

// TestCreateLab_InvalidData tests creating lab with invalid data
func (s *LabsTestSuite) TestCreateLab_InvalidData() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Test Lab",
		"course_id":    "invalid-uuid",
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/labs"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestCreateLab_MissingFields tests creating lab with missing fields
func (s *LabsTestSuite) TestCreateLab_MissingFields() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Test Lab",
		// Missing course_id
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/labs"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestGetLab_Admin_Success tests admin getting lab details
func (s *LabsTestSuite) TestGetLab_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs/"+labID), nil, adminToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().Equal(labID, result["id"])
}

// TestGetLab_NotFound tests getting non-existent lab
func (s *LabsTestSuite) TestGetLab_NotFound() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs/nonexistent-lab-id"), nil, adminToken)

	s.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

// TestUpdateLab_Admin_Success tests admin updating a lab
func (s *LabsTestSuite) TestUpdateLab_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Updated Lab Name",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/labs/"+labID), reqBody, adminToken)

	s.AssertSuccess(resp)
}

// TestUpdateLab_Instructor_OwnLab tests instructor updating their own lab
func (s *LabsTestSuite) TestUpdateLab_Instructor_OwnLab() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Instructor.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Instructor.UserID)
	defer s.CleanupTestLab(labID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Updated by Instructor",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/labs/"+labID), reqBody, instructorToken)

	s.AssertSuccess(resp)
}

// TestUpdateLab_InvalidData tests updating lab with invalid data
func (s *LabsTestSuite) TestUpdateLab_InvalidData() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"course_id": "invalid-uuid",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/cms/labs/"+labID), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestDeleteLab_Admin_Success tests admin deleting a lab
func (s *LabsTestSuite) TestDeleteLab_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("DELETE", BuildURL("/cms/labs/"+labID), nil, adminToken)

	s.AssertSuccess(resp)

	// Verify lab is soft deleted
	var isDeleted bool
	err := s.DB.Get(&isDeleted, "SELECT is_deleted FROM labs WHERE id = $1", labID)
	s.Require().NoError(err)
	s.Assert().True(isDeleted)
}

// TestDeleteLab_Instructor_OwnLab tests instructor deleting their own lab
func (s *LabsTestSuite) TestDeleteLab_Instructor_OwnLab() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Instructor.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Instructor.UserID)

	instructorToken := s.GenerateTestJWT(s.TestUser.Instructor.UserID, s.TestUser.Instructor.Username, s.TestUser.Instructor.Roles)

	resp := s.RequestWithAuth("DELETE", BuildURL("/cms/labs/"+labID), nil, instructorToken)

	s.AssertSuccess(resp)
}

// TestListLabSections_Admin_Success tests admin listing lab sections
func (s *LabsTestSuite) TestListLabSections_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	semesterID := s.CreateTestSemester()
	sectionID := s.CreateTestSection(courseID, semesterID, []string{s.TestUser.Instructor.UserID}, []string{s.TestUser.Student.UserID})
	defer s.CleanupTestSection(sectionID)

	s.CreateTestLabSection(labID, sectionID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs/"+labID+"/sections"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestListLabMaterials_Admin_Success tests admin listing lab materials
func (s *LabsTestSuite) TestListLabMaterials_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	// Create a material for the lab
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), s.TestUser.Admin.UserID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs/"+labID+"/materials"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestGetAllLabMaterials_Admin_Success tests admin getting all lab materials
func (s *LabsTestSuite) TestGetAllLabMaterials_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	// Create materials for the lab
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), s.TestUser.Admin.UserID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/cms/labs/"+labID+"/materials/all"), nil, adminToken)

	s.AssertSuccess(resp)

	var result []interface{}
	s.ParseJSONResponse(resp, &result)

	// Cleanup
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestAddMaterialToLab_Admin_Success tests admin adding material to lab
func (s *LabsTestSuite) TestAddMaterialToLab_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	// Create a material first (standalone, not associated with lab yet)
	materialID := s.CreateTestMaterialStandalone("code", s.TestUser.Admin.UserID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"material_id": materialID,
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/labs/"+labID+"/materials"), reqBody, adminToken)

	s.AssertCreated(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM lab_materials WHERE material_id = $1", materialID)
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}

// TestRemoveMaterialFromLab_Admin_Success tests admin removing material from lab
func (s *LabsTestSuite) TestRemoveMaterialFromLab_Admin_Success() {
	// Create test data
	courseID := s.CreateTestCourse(s.TestUser.Admin.UserID)
	defer s.CleanupTestCourse(courseID)

	labID := s.CreateTestLab(courseID, s.TestUser.Admin.UserID)
	defer s.CleanupTestLab(labID)

	// Create a material for the lab
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), s.TestUser.Admin.UserID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"material_id": materialID,
	}

	resp := s.RequestWithAuth("POST", BuildURL("/cms/labs/"+labID+"/materials/delete"), reqBody, adminToken)

	s.AssertSuccess(resp)

	// Cleanup
	s.DB.Exec("DELETE FROM materials WHERE id = $1", materialID)
}
