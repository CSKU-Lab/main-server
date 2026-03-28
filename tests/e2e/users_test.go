//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

// UsersTestSuite tests user management routes
// Routes tested:
// - GET /admin/users - List users (admin only)
// - POST /admin/users - Create user (admin only)
// - GET /users/:userID - Get user profile
// - PATCH /admin/users/:userID - Update user (admin only)
// - DELETE /admin/users/:userID - Delete user (admin only)
// - POST /admin/users/deleteMany - Delete multiple users (admin only)
type UsersTestSuite struct {
	TestSuite
}

func TestUsersTestSuite(t *testing.T) {
	CheckE2EEnabled(t)
	suite.Run(t, new(UsersTestSuite))
}

// TestListUsers_Admin_Success tests admin listing users
func (s *UsersTestSuite) TestListUsers_Admin_Success() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/admin/users"), nil, adminToken)

	s.AssertSuccess(resp)
	s.AssertPagination(resp)
}

// TestListUsers_WithPagination tests listing users with pagination params
func (s *UsersTestSuite) TestListUsers_WithPagination() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/admin/users?page=1&page_size=5"), nil, adminToken)

	s.AssertSuccess(resp)

	// Verify pagination data
	pagination := s.GetNestedResponseField(resp, "pagination")
	s.Assert().NotNil(pagination)
}

// TestListUsers_WithSearch tests listing users with search
func (s *UsersTestSuite) TestListUsers_WithSearch() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/admin/users?search=test"), nil, adminToken)

	s.AssertSuccess(resp)
}

// TestListUsers_NonAdmin_Forbidden tests non-admin trying to list users
func (s *UsersTestSuite) TestListUsers_NonAdmin_Forbidden() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/admin/users"), nil, studentToken)

	s.AssertForbidden(resp)
}

// TestListUsers_Unauthorized tests listing users without auth
func (s *UsersTestSuite) TestListUsers_Unauthorized() {
	resp := s.RequestWithoutAuth("GET", BuildURL("/admin/users"), nil)

	s.AssertUnauthorized(resp)
}

// TestCreateUser_Admin_Success tests admin creating a new user
func (s *UsersTestSuite) TestCreateUser_Admin_Success() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"username":     "new_test_user_" + generateRandomString(6),
		"display_name": "New Test User",
		"roles":        []string{"student"},
		"type":         "credential",
		"password":     "SecurePass123!",
		"email":        "newuser@example.com",
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users"), reqBody, adminToken)

	s.AssertCreated(resp)
}

// TestCreateUser_InvalidEmail tests creating user with invalid email
func (s *UsersTestSuite) TestCreateUser_InvalidEmail() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"username":     "testuser_invalid_email",
		"display_name": "Test User",
		"roles":        []string{"student"},
		"type":         "credential",
		"password":     "SecurePass123!",
		"email":        "invalid-email",
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestCreateUser_ShortPassword tests creating user with short password
func (s *UsersTestSuite) TestCreateUser_ShortPassword() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"username":     "testuser_short_pass",
		"display_name": "Test User",
		"roles":        []string{"student"},
		"type":         "credential",
		"password":     "short",
		"email":        "test@example.com",
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestCreateUser_MissingRequiredFields tests creating user without required fields
func (s *UsersTestSuite) TestCreateUser_MissingRequiredFields() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Test User",
		// Missing username and roles
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestCreateUser_NonAdmin_Forbidden tests non-admin trying to create user
func (s *UsersTestSuite) TestCreateUser_NonAdmin_Forbidden() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"username":     "unauthorized_user",
		"display_name": "Unauthorized User",
		"roles":        []string{"student"},
		"type":         "credential",
		"password":     "SecurePass123!",
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users"), reqBody, studentToken)

	s.AssertForbidden(resp)
}

// TestGetUser_Admin_Success tests admin getting any user profile
func (s *UsersTestSuite) TestGetUser_Admin_Success() {
	// Create a test user to retrieve
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/users/"+userID), nil, adminToken)

	s.AssertSuccess(resp)

	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().Equal(userID, result["id"])
}

// TestGetUser_OwnProfile tests user getting their own profile
func (s *UsersTestSuite) TestGetUser_OwnProfile() {
	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/users/"+s.TestUser.Student.UserID), nil, studentToken)

	s.AssertSuccess(resp)
}

// TestGetUser_NotFound tests getting non-existent user
func (s *UsersTestSuite) TestGetUser_NotFound() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("GET", BuildURL("/users/nonexistent-user-id"), nil, adminToken)

	s.Assert().Equal(http.StatusInternalServerError, resp.StatusCode)
}

// TestGetUser_Unauthorized tests getting user without auth
func (s *UsersTestSuite) TestGetUser_Unauthorized() {
	resp := s.RequestWithoutAuth("GET", BuildURL("/users/some-user-id"), nil)

	s.AssertUnauthorized(resp)
}

// TestUpdateUser_Admin_Success tests admin updating a user
func (s *UsersTestSuite) TestUpdateUser_Admin_Success() {
	// Create a test user to update
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Updated Name",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/admin/users/"+userID), reqBody, adminToken)

	s.Assert().Equal(http.StatusAccepted, resp.StatusCode)
}

// TestUpdateUser_InvalidData tests updating user with invalid data
func (s *UsersTestSuite) TestUpdateUser_InvalidData() {
	// Create a test user to update
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"email": "invalid-email",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/admin/users/"+userID), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestUpdateUser_NonAdmin_Forbidden tests non-admin trying to update user
func (s *UsersTestSuite) TestUpdateUser_NonAdmin_Forbidden() {
	// Create a test user to update
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	reqBody := map[string]interface{}{
		"display_name": "Hacked Name",
	}

	resp := s.RequestWithAuth("PATCH", BuildURL("/admin/users/"+userID), reqBody, studentToken)

	s.AssertForbidden(resp)
}

// TestDeleteUser_Admin_Success tests admin deleting a user
func (s *UsersTestSuite) TestDeleteUser_Admin_Success() {
	// Create a test user to delete
	userID := s.CreateTestUser("student", []string{"student"})

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	resp := s.RequestWithAuth("DELETE", BuildURL("/admin/users/"+userID), nil, adminToken)

	s.AssertNoContent(resp)

	// Verify user is deleted
	var count int
	err := s.DB.Get(&count, "SELECT COUNT(*) FROM users WHERE id = $1 AND is_deleted = false", userID)
	s.Require().NoError(err)
	s.Assert().Equal(0, count)
}

// TestDeleteUser_NonAdmin_Forbidden tests non-admin trying to delete user
func (s *UsersTestSuite) TestDeleteUser_NonAdmin_Forbidden() {
	// Create a test user to delete
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	studentToken := s.GenerateTestJWT(s.TestUser.Student.UserID, s.TestUser.Student.Username, s.TestUser.Student.Roles)

	resp := s.RequestWithAuth("DELETE", BuildURL("/admin/users/"+userID), nil, studentToken)

	s.AssertForbidden(resp)
}

// TestDeleteManyUsers_Admin_Success tests admin deleting multiple users
func (s *UsersTestSuite) TestDeleteManyUsers_Admin_Success() {
	// Create test users to delete
	userID1 := s.CreateTestUser("student", []string{"student"})
	userID2 := s.CreateTestUser("student", []string{"student"})

	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"ids": []string{userID1, userID2},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users/deleteMany"), reqBody, adminToken)

	s.AssertNoContent(resp)
}

// TestDeleteManyUsers_InvalidIDs tests deleting users with invalid IDs
func (s *UsersTestSuite) TestDeleteManyUsers_InvalidIDs() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"ids": []string{"invalid-id-1", "invalid-id-2"},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users/deleteMany"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestDeleteManyUsers_EmptyList tests deleting users with empty list
func (s *UsersTestSuite) TestDeleteManyUsers_EmptyList() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"ids": []string{},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users/deleteMany"), reqBody, adminToken)

	s.AssertBadRequest(resp)
}

// TestImportUsers_Admin_Success tests admin importing multiple users
func (s *UsersTestSuite) TestImportUsers_Admin_Success() {
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	reqBody := map[string]interface{}{
		"users": []map[string]interface{}{
			{
				"username":     "imported_user_1_" + generateRandomString(6),
				"display_name": "Imported User 1",
				"roles":        []string{"student"},
				"type":         "credential",
				"password":     "SecurePass123!",
			},
			{
				"username":     "imported_user_2_" + generateRandomString(6),
				"display_name": "Imported User 2",
				"roles":        []string{"student"},
				"type":         "credential",
				"password":     "SecurePass123!",
			},
		},
	}

	resp := s.RequestWithAuth("POST", BuildURL("/admin/users/import"), reqBody, adminToken)

	s.AssertCreated(resp)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
