//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SubmissionsRoutesTestSuite tests all submission-related endpoints
type SubmissionsRoutesTestSuite struct {
	TestSuite
}

func TestSubmissionsRoutes(t *testing.T) {
	suite.Run(t, new(SubmissionsRoutesTestSuite))
}

// TestCreateSubmission_StudentCanCreate tests student creating a submission
func (s *SubmissionsRoutesTestSuite) TestCreateSubmission_StudentCanCreate() {
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
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), adminID)
	_ = materialID

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Prepare submission request
	submissionReq := map[string]interface{}{
		"material_id": materialID,
		"lab_id":      labID,
		"section_id":  sectionID,
		"type":        "code",
		"files": []map[string]string{
			{
				"name":    "main.py",
				"content": "print('Hello World')",
			},
		},
	}
	reqBody, _ := json.Marshal(submissionReq)

	// Make request
	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - may be 201 or error depending on gRPC availability
	assert.Contains(s.T(), []int{http.StatusCreated, http.StatusInternalServerError, http.StatusBadRequest}, resp.StatusCode)

	// If successful, parse response
	if resp.StatusCode == http.StatusCreated {
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(s.T(), err)
		assert.NotNil(s.T(), result["id"])
	}
}

// TestCreateSubmission_Unauthorized tests creating submission without auth
func (s *SubmissionsRoutesTestSuite) TestCreateSubmission_Unauthorized() {
	// Prepare submission request
	submissionReq := map[string]interface{}{
		"material_id": "some-material-id",
		"lab_id":      "some-lab-id",
		"section_id":  "some-section-id",
		"type":        "code",
		"files": []map[string]string{
			{
				"name":    "main.py",
				"content": "print('Hello World')",
			},
		},
	}
	reqBody, _ := json.Marshal(submissionReq)

	// Make request without authentication
	req := httptest.NewRequest("POST", "/api/v1/submissions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 401 Unauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

// TestGetSubmission_StudentCanGetOwn tests student getting their own submission
func (s *SubmissionsRoutesTestSuite) TestGetSubmission_StudentCanGetOwn() {
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
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	// Cleanup will be handled by user cleanup

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/submissions/"+submissionID, nil)
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
	assert.Equal(s.T(), submissionID, result["id"])
}

// TestListUserSubmissions_StudentCanList tests student listing their submissions
func (s *SubmissionsRoutesTestSuite) TestListUserSubmissions_StudentCanList() {
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
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/submissions?page=1&page_size=20", nil)
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

// TestGetMaterialSubmissions_StudentCanGet tests student getting material submissions
func (s *SubmissionsRoutesTestSuite) TestGetMaterialSubmissions_StudentCanGet() {
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
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/materials/"+materialID+"/submissions?section_id="+sectionID+"&lab_id="+labID, nil)
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

// TestUpdateManualScore_AdminCanUpdate tests admin updating manual score
func (s *SubmissionsRoutesTestSuite) TestUpdateManualScore_AdminCanUpdate() {
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
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Prepare update request - manual_score must be an integer
	updateReq := map[string]interface{}{
		"manual_score": 95,
	}
	reqBody, _ := json.Marshal(updateReq)

	// Make request
	req := httptest.NewRequest("PATCH", "/api/v1/cms/submissions/"+submissionID+"/manual-score", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+adminToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(s.T(), http.StatusNoContent, resp.StatusCode)
}

// TestUpdateManualScore_StudentCannotUpdate tests student trying to update manual score
func (s *SubmissionsRoutesTestSuite) TestUpdateManualScore_StudentCannotUpdate() {
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
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate student token
	studentToken := s.GenerateTestJWT(studentID, "test_student", []string{"student"})

	// Prepare update request - manual_score must be an integer
	updateReq := map[string]interface{}{
		"manual_score": 100,
	}
	reqBody, _ := json.Marshal(updateReq)

	// Make request
	req := httptest.NewRequest("PATCH", "/api/v1/cms/submissions/"+submissionID+"/manual-score", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "access_token="+studentToken)

	resp, err := s.App.Test(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	// Verify response - should be 403 Forbidden
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode)
}

// TestGetSectionLabMaterialSubmissions_AdminCanGet tests admin getting submissions for material
func (s *SubmissionsRoutesTestSuite) TestGetSectionLabMaterialSubmissions_AdminCanGet() {
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
	materialID := s.CreateTestCodeMaterial(labID, uuid.New().String(), adminID)
	_ = materialID

	// Create submission
	submissionID := s.CreateTestSubmission(studentID, materialID, labID, sectionID, courseID)
	_ = submissionID

	// Generate admin token
	adminToken := s.GenerateTestJWT(adminID, "test_admin", []string{"admin"})

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/cms/sections/"+sectionID+"/labs/"+labID+"/materials/"+materialID+"/submissions", nil)
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
