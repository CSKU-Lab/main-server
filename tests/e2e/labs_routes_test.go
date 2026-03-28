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
	"github.com/stretchr/testify/suite"
)

// LabsRoutesTestSuite tests all lab management endpoints
type LabsRoutesTestSuite struct {
	TestSuite
}

func TestLabsRoutes(t *testing.T) {
	suite.Run(t, new(LabsRoutesTestSuite))
}

// TestListLabs_AdminCanList tests admin listing labs
func (s *LabsRoutesTestSuite) TestListLabs_AdminCanList() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/labs?page=1&page_size=10", nil)
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

// TestCreateLab_AdminCanCreate tests admin creating a lab
func (s *LabsRoutesTestSuite) TestCreateLab_AdminCanCreate() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare create lab request
	createReq := map[string]interface{}{
		"display_name": "E2E Test Lab",
		"course_id":    courseID,
	}
	reqBody, _ := json.Marshal(createReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/labs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
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
	if labID, ok := result["id"].(string); ok {
		s.CleanupTestLab(labID)
	}
}

// TestCreateLab_InstructorCanCreate tests instructor creating a lab
func (s *LabsRoutesTestSuite) TestCreateLab_InstructorCanCreate() {
	// Create instructor user
	instructorID := s.CreateTestUser("instructor", []string{"instructor"})
	defer s.CleanupTestUser(instructorID)

	// Create a test course
	courseID := s.CreateTestCourse(instructorID)
	defer s.CleanupTestCourse(courseID)

	// Generate instructor token
	instructorToken := s.GenerateTestJWT(instructorID, "test_instructor", []string{"instructor"})

	// Prepare create lab request
	createReq := map[string]interface{}{
		"display_name": "E2E Test Lab by Instructor",
		"course_id":    courseID,
	}
	reqBody, _ := json.Marshal(createReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/labs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
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
	if labID, ok := result["id"].(string); ok {
		s.CleanupTestLab(labID)
	}
}

// TestCreateLab_StudentCannotCreate tests student trying to create lab
func (s *LabsRoutesTestSuite) TestCreateLab_StudentCannotCreate() {
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

	// Prepare create lab request
	createReq := map[string]interface{}{
		"display_name": "E2E Test Lab by Student",
		"course_id":    courseID,
	}
	reqBody, _ := json.Marshal(createReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/labs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 403 Forbidden
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

// TestGetLabByID_Success tests getting a lab by ID
func (s *LabsRoutesTestSuite) TestGetLabByID_Success() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/labs/"+labID, nil)
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
	assert.Equal(s.T(), labID, result["id"])
}

// TestUpdateLab_AdminCanUpdate tests admin updating a lab
func (s *LabsRoutesTestSuite) TestUpdateLab_AdminCanUpdate() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare update request
	updateReq := map[string]interface{}{
		"display_name": "Updated Lab Name",
	}
	reqBody, _ := json.Marshal(updateReq)

	// Make request
	req := httptest.NewRequest("PATCH", "/api/v1/cms/labs/"+labID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// TestDeleteLab_AdminCanDelete tests admin deleting a lab
func (s *LabsRoutesTestSuite) TestDeleteLab_AdminCanDelete() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	labID := s.CreateTestLab(courseID, adminID)
	// Don't defer cleanup - we're testing deletion

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("DELETE", "/api/v1/cms/labs/"+labID, nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// TestGetLabSections_AdminCanGet tests admin getting lab sections
func (s *LabsRoutesTestSuite) TestGetLabSections_AdminCanGet() {
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
	req := httptest.NewRequest("GET", "/api/v1/cms/labs/"+labID+"/sections", nil)
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

// TestAddMaterialToLab_AdminCanAdd tests admin adding material to lab
func (s *LabsRoutesTestSuite) TestAddMaterialToLab_AdminCanAdd() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create a material
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	// Cleanup will be handled by lab cleanup

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare request
	addReq := map[string]interface{}{
		"material_id": materialID,
		"position":    1,
	}
	reqBody, _ := json.Marshal(addReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/cms/labs/"+labID+"/materials", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
}

// TestListLabMaterials_AdminCanList tests admin listing lab materials
func (s *LabsRoutesTestSuite) TestListLabMaterials_AdminCanList() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create a material
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/labs/"+labID+"/materials", nil)
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

// TestGetAllLabMaterials_AdminCanGet tests admin getting all lab materials
func (s *LabsRoutesTestSuite) TestGetAllLabMaterials_AdminCanGet() {
	// Create admin user
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)

	// Create a test course and lab
	courseID := s.CreateTestCourse(adminID)
	defer s.CleanupTestCourse(courseID)
	labID := s.CreateTestLab(courseID, adminID)
	defer s.CleanupTestLab(labID)

	// Create materials
	materialID1 := s.CreateTestMaterial(labID, "code", adminID)
	materialID2 := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID1
	_ = materialID2

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/labs/"+labID+"/materials/all", nil)
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

// TestStudentGetLab_StudentCanGet tests student getting lab details
func (s *LabsRoutesTestSuite) TestStudentGetLab_StudentCanGet() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, and lab
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

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Prepare request body
	reqBody := map[string]interface{}{
		"section_id": sectionID,
	}
	body, _ := json.Marshal(reqBody)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/labs/"+labID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
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
	assert.NotNil(s.T(), result["id"])
}

// TestStudentListLabMaterials_StudentCanList tests student listing lab materials
func (s *LabsRoutesTestSuite) TestStudentListLabMaterials_StudentCanList() {
	// Create admin and student users
	adminID := s.CreateTestUser("admin", []string{"admin"})
	defer s.CleanupTestUser(adminID)
	studentID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(studentID)

	// Create a test course, section, and lab
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

	// Create materials
	materialID := s.CreateTestMaterial(labID, "code", adminID)
	_ = materialID

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/labs/"+labID+"/materials?section_id="+sectionID, nil)
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
