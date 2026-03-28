//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

// AuthTestSuite tests authentication routes
// Routes tested:
// - POST /auth/sign-in/credential - Login with username/password
// - GET /auth/sign-in/google - Google OAuth login (not tested - requires external service)
// - GET /auth/sign-in/google/callback - Google OAuth callback (not tested - requires external service)
// - POST /auth/refresh-token - Refresh access token
type AuthTestSuite struct {
	TestSuite
}

func TestAuthTestSuite(t *testing.T) {
	CheckE2EEnabled(t)
	suite.Run(t, new(AuthTestSuite))
}

// TestCredentialLogin_Success tests successful login with valid credentials
func (s *AuthTestSuite) TestCredentialLogin_Success() {
	// Create a test user with known credentials
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	// Set password for the user
	_, err := s.DB.ExecContext(s.Ctx, `
		UPDATE user_passwords 
		SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQzBZN0UfGNEJH0O9Q.DqL5/7E.S'
		WHERE user_id = $1
	`, userID)
	s.Require().NoError(err)

	reqBody := map[string]string{
		"username": s.TestUser.Student.Username,
		"password": "TestPassword123!",
	}

	resp := s.RequestWithoutAuth("POST", BuildURL("/auth/sign-in/credential"), reqBody)

	s.AssertSuccess(resp)

	// Verify response contains expected fields
	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	s.Assert().Equal("success", result["message"])
}

// TestCredentialLogin_WrongPassword tests login with incorrect password
func (s *AuthTestSuite) TestCredentialLogin_WrongPassword() {
	// Create a test user
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	reqBody := map[string]string{
		"username": s.TestUser.Student.Username,
		"password": "WrongPassword123!",
	}

	resp := s.RequestWithoutAuth("POST", BuildURL("/auth/sign-in/credential"), reqBody)

	s.AssertUnauthorized(resp)
	s.AssertErrorResponse(resp, "unauthorized")
}

// TestCredentialLogin_UserNotFound tests login with non-existent user
func (s *AuthTestSuite) TestCredentialLogin_UserNotFound() {
	reqBody := map[string]string{
		"username": "nonexistent_user_12345",
		"password": "SomePassword123!",
	}

	resp := s.RequestWithoutAuth("POST", BuildURL("/auth/sign-in/credential"), reqBody)

	s.AssertUnauthorized(resp)
	s.AssertErrorResponse(resp, "unauthorized")
}

// TestCredentialLogin_MissingCredentials tests login with missing credentials
func (s *AuthTestSuite) TestCredentialLogin_MissingCredentials() {
	// Missing password
	reqBody := map[string]string{
		"username": "testuser",
	}

	resp := s.RequestWithoutAuth("POST", BuildURL("/auth/sign-in/credential"), reqBody)

	s.AssertBadRequest(resp)
}

// TestCredentialLogin_MissingUsername tests login with missing username
func (s *AuthTestSuite) TestCredentialLogin_MissingUsername() {
	reqBody := map[string]string{
		"password": "SomePassword123!",
	}

	resp := s.RequestWithoutAuth("POST", BuildURL("/auth/sign-in/credential"), reqBody)

	s.AssertBadRequest(resp)
}

// TestRefreshToken_Success tests successful token refresh with valid refresh token
func (s *AuthTestSuite) TestRefreshToken_Success() {
	// Create a test user and generate tokens
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	accessToken := s.GenerateTestJWT(userID, "testuser", []string{"student"})
	refreshToken := s.GenerateTestRefreshToken(userID)
	s.StoreRefreshToken(userID, refreshToken)

	// Make request with cookies
	req := httptest.NewRequest("POST", BuildURL("/auth/refresh-token"), nil)
	req.Header.Set("Cookie", "access_token="+accessToken+"; refresh_token="+refreshToken)

	resp, err := s.App.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)
}

// TestRefreshToken_InvalidToken tests refresh with invalid token
func (s *AuthTestSuite) TestRefreshToken_InvalidToken() {
	req := httptest.NewRequest("POST", BuildURL("/auth/refresh-token"), nil)
	req.Header.Set("Cookie", "access_token=invalid_token; refresh_token=invalid_token")

	resp, err := s.App.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode)
}

// TestRefreshToken_NoCookies tests refresh without any cookies
func (s *AuthTestSuite) TestRefreshToken_NoCookies() {
	resp := s.RequestWithoutAuth("POST", BuildURL("/auth/refresh-token"), nil)

	s.AssertUnauthorized(resp)
}

// TestRefreshToken_ExpiredAccessToken tests refresh with expired access token but valid refresh token
func (s *AuthTestSuite) TestRefreshToken_ExpiredAccessToken() {
	// Create a test user
	userID := s.CreateTestUser("student", []string{"student"})
	defer s.CleanupTestUser(userID)

	// Generate an expired access token (we'll use a dummy one)
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZXhwIjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	refreshToken := s.GenerateTestRefreshToken(userID)
	s.StoreRefreshToken(userID, refreshToken)

	req := httptest.NewRequest("POST", BuildURL("/auth/refresh-token"), nil)
	req.Header.Set("Cookie", "access_token="+accessToken+"; refresh_token="+refreshToken)

	resp, err := s.App.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Should succeed because refresh token is valid
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
}

// TestProtectedRoute_WithoutAuth tests accessing protected route without authentication
func (s *AuthTestSuite) TestProtectedRoute_WithoutAuth() {
	// Try to access a protected endpoint
	resp := s.RequestWithoutAuth("GET", BuildURL("/admin/users"), nil)

	s.AssertUnauthorized(resp)
}

// TestProtectedRoute_WithInvalidToken tests accessing protected route with invalid token
func (s *AuthTestSuite) TestProtectedRoute_WithInvalidToken() {
	resp := s.RequestWithAuth("GET", BuildURL("/admin/users"), nil, "invalid_token")

	s.AssertUnauthorized(resp)
}

// TestProtectedRoute_WithValidToken tests accessing protected route with valid token
func (s *AuthTestSuite) TestProtectedRoute_WithValidToken() {
	// Create admin user and generate token
	adminToken := s.GenerateTestJWT(s.TestUser.Admin.UserID, s.TestUser.Admin.Username, s.TestUser.Admin.Roles)

	// This will likely return other errors (like no users found) but should not be 401
	resp := s.RequestWithAuth("GET", BuildURL("/admin/users"), nil, adminToken)

	// Should not be unauthorized
	s.Assert().NotEqual(http.StatusUnauthorized, resp.StatusCode)
}
